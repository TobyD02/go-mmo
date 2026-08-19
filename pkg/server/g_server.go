package server

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type GServer struct {
	Clients              map[string]*GServerClient
	WorldController      *GWorldController
	clientsMutex         sync.RWMutex
	queuedConnections    []string
	queuedDisconnections []string
	upgrader             websocket.Upgrader
	TickSpeed            time.Duration
	tick                 int
}

func NewGServer(tickSpeed time.Duration, worldWidth int, worldHeight int) *GServer {
	clients := make(map[string]*GServerClient)

	server := &GServer{
		Clients:              clients,
		queuedConnections:    make([]string, 0),
		queuedDisconnections: make([]string, 0),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		tick:      0,
		TickSpeed: tickSpeed,
	}

	worldController := NewGWorldController(worldWidth, worldHeight, func() int {
		return server.tick
	})
	worldController.SetupWorld(true)

	server.WorldController = worldController
	return server
}

func (s *GServer) HandleClientConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade to a websocket connection
	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	defer conn.Close()

	client := NewGServerClient(conn)
	err = client.EstablishConnection()
	if err != nil {
		log.Printf("failed to establish client connection: %s", err)
		return
	}

	s.clientsMutex.Lock()

	if _, exists := s.Clients[client.ID]; exists {
		s.clientsMutex.Unlock()
		client.WriteMessage(
			websocket.TextMessage,
			[]byte("Client already connected"),
		)
		return
	}

	// Assign the client connection

	s.Clients[client.ID] = client
	s.queuedConnections = append(s.queuedConnections, client.ID)

	s.clientsMutex.Unlock()

	// Defer deletion on disconnect
	defer s.removeClient(client)

	// Initialise the ping loop (ensure client is connected)
	go client.PingLoop()
	// log.Println("client connected: ", client.ID)

	err = client.ReadMessages()
	if err != nil {
		// fmt.Printf("client disconnected %s", client.ID)
	}
}

func (s *GServer) buildInitialWorldStateMessage() ([]byte, error) {
	// First message is world state
	serverInitialWorldStateMessage, err := messages.NewGServerInitialWorldStateMessage(s.WorldController.GameWorld)
	if err != nil {
		return nil, fmt.Errorf("Failed to get initial world state")
	}

	// Marshal the message
	return json.Marshal(serverInitialWorldStateMessage)
}

func (s *GServer) sendInitialWorldState(client *GServerClient, initialWorldStateMessage []byte) {
	// Send the inital message
	err := client.WriteMessage(websocket.TextMessage, initialWorldStateMessage)
	if err != nil {
		log.Printf("Failed to send initial world state")
	}
}

func (s *GServer) removeClient(client *GServerClient) {
	s.clientsMutex.Lock()

	delete(s.Clients, client.ID)
	s.queuedDisconnections = append(s.queuedDisconnections, client.ID)
	s.clientsMutex.Unlock()

	client.Conn.Close()
}

func (s *GServer) clientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	return len(s.Clients)
}

func (s *GServer) GameLoop() {
	ticker := time.NewTicker(s.TickSpeed)
	defer ticker.Stop()

	peakTickSpeed := time.Since(time.Now())

	for range ticker.C {
		now := time.Now()
		s.doTick()

		timeTaken := time.Since(now)
		if timeTaken > peakTickSpeed {
			peakTickSpeed = timeTaken
		}
		if timeTaken >= time.Millisecond*5 {
			log.Printf(
				"TICK | slowest: %v, took: %v, target: %v, clients: %v",
				peakTickSpeed,
				timeTaken,
				s.TickSpeed,
				s.clientCount(),
			)
		}
	}
}

func (s *GServer) snapshotClients() map[string]*GServerClient {
	s.clientsMutex.RLock()
	currentClients := make(map[string]*GServerClient, len(s.Clients))
	maps.Copy(currentClients, s.Clients)
	s.clientsMutex.RUnlock()
	return currentClients
}

func (s *GServer) doTick() {
	s.tick++

	// Handle connections and disconnections first
	s.handleClientConnectionsAndDisconnections()

	// Get snapshot of current clients
	currentClients := s.snapshotClients()

	// Do the game world tick
	s.doGameWorldTick(currentClients)

	// Build diff from changes and send to clients
	diff := s.WorldController.BuildWorldDiff()
	err := s.relayWorldDiff(&diff, currentClients)
	if err != nil {
		log.Printf("Failed to relay world diff: %s", err)
	}
}

func (s *GServer) relayWorldDiff(worldDiff *game.GameWorldDiff, currentClients map[string]*GServerClient) error {
	msg, err := messages.NewGServerWorldDiffMessage(worldDiff)
	if err != nil {
		return fmt.Errorf("failed to generate world diff message")
	}

	payload, err := json.Marshal(msg)

	if err != nil {
		return fmt.Errorf("Failed to generate world diff message: %s", err)
	}

	// Relay updates to clients
	for _, client := range currentClients {
		client.WriteMessage(websocket.TextMessage, payload)
	}

	return nil
}

func (s *GServer) doGameWorldTick(currentClients map[string]*GServerClient) {
	// Do Client Actions
	for clientID, client := range currentClients {
		clientMessages := client.DrainMessages()
		s.handleMessages(clientID, clientMessages) // Client actions affect new game world
	}

	// Run tickers last
	s.WorldController.DoTickers()
}

func (s *GServer) getClient(clientID string) *GServerClient {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	return s.Clients[clientID]
}

func (s *GServer) handleClientConnectionsAndDisconnections() {
	s.clientsMutex.Lock()

	connections := s.queuedConnections
	disconnections := s.queuedDisconnections

	s.queuedConnections = nil
	s.queuedDisconnections = nil

	s.clientsMutex.Unlock()

	if len(connections) > 0 {
		initialWorldStateMessage, err := s.buildInitialWorldStateMessage()
		if err != nil {
			log.Printf("Failed to build initial world state message - who knows what will happend: %s", err)
		}

		// Handle new connections
		for _, clientID := range connections {

			client := s.getClient(clientID)

			if client == nil {
				continue
			}

			err := s.WorldController.SpawnNewPlayer(clientID)
			if err != nil {
				s.removeClient(s.Clients[clientID])
				continue
			}

			s.sendInitialWorldState(client, initialWorldStateMessage)
		}
	}

	// Handle queuedDisconnections
	for _, clientID := range disconnections {
		s.WorldController.DeletePlayer(clientID)
	}
}

// @TODO - need a better way to handle this. Too much logic in the server
// Maybe some message processor - Cannot modify the world directly because i need to account for state?
// Or just process the messages - when collecting only one of each message type (the latest) can be used
func (s *GServer) handleMessages(
	clientID string,
	msgs map[messages.GMessageType]*messages.GMessage,
) {
	for messageType, message := range msgs {
		switch messageType {

		case messages.TClientMoveMessage:
			moveData, err := messages.ParseGClientMoveMessageData(message.Data)
			if err != nil {
				log.Printf("error: failed to parse move data")
			}

			s.WorldController.MovePlayer(clientID, moveData.Dx, moveData.Dy)

		case messages.TClientInteractMessage:
			interactData, err := messages.ParseGClientInteractMessageData(message.Data)
			if err != nil {
				log.Printf("error: failed to parse move data")
			}

			s.WorldController.InteractWith(clientID, interactData.InteractableID)
		}

	}

}
