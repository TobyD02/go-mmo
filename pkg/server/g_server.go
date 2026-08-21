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
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type GServer struct {
	Clients      map[string]*GServerClient
	clientsMutex sync.RWMutex

	ClientsReadOnly      map[string]*GServerClient
	clientsReadOnlyMutex sync.RWMutex

	WorldController *GWorldController

	queuedConnections    []string
	queuedDisconnections []string

	upgrader websocket.Upgrader

	TickSpeed time.Duration
	tick      int

	MessageRouter *GMessageRouter
}

func NewGServer(tickSpeed time.Duration, worldWidth int, worldHeight int) *GServer {
	server := &GServer{
		Clients:         make(map[string]*GServerClient),
		ClientsReadOnly: make(map[string]*GServerClient),

		queuedConnections:    make([]string, 0),
		queuedDisconnections: make([]string, 0),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		tick:          0,
		TickSpeed:     tickSpeed,
		MessageRouter: NewGMessageRouter(),
	}

	worldController := NewGWorldController(
		worldWidth,
		worldHeight,
		func() int {
			return server.tick
		},
		func() *GMessageRouter {
			return server.MessageRouter
		},
	)
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

	client := NewGServerClient(conn)
	err = client.EstablishConnection()
	if err != nil {
		log.Printf("failed to establish client connection: %s", err)

		client.Close()
		return
	}

	s.clientsMutex.Lock()

	if _, exists := s.Clients[client.ID]; exists {
		s.clientsMutex.Unlock()

		log.Printf("client already connected: %s", client.ID)

		return
	}

	// Assign the client connection

	s.Clients[client.ID] = client
	s.queuedConnections = append(s.queuedConnections, client.ID)

	s.clientsMutex.Unlock()

	// Defer deletion on disconnect
	defer s.removeClient(client)

	// Initialise the ping loop (ensure client is connected) and write loop (for outbound messages)
	go client.WriteLoop()
	go client.PingLoop()
	// log.Println("client connected: ", client.ID)

	err = client.ReadLoop()
	if err != nil {
		fmt.Printf("client disconnected %s", client.ID)
	}
}

func (s *GServer) removeClient(client *GServerClient) {
	s.clientsMutex.Lock()

	delete(s.Clients, client.ID)
	s.queuedDisconnections = append(s.queuedDisconnections, client.ID)
	s.clientsMutex.Unlock()

	client.Close()
}

func (s *GServer) HandleClientConnectionReadOnly(w http.ResponseWriter, r *http.Request) {
	// Upgrade to a websocket connection
	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	client := NewGServerClient(conn)
	err = client.EstablishConnection()
	if err != nil {
		log.Printf("failed to establish client connection: %s", err)

		client.Close()
		return
	}

	s.clientsReadOnlyMutex.Lock()

	if _, exists := s.ClientsReadOnly[client.ID]; exists {
		s.clientsReadOnlyMutex.Unlock()

		log.Printf("(read only) client already connected: %s", client.ID)

		return
	}

	// Assign the client connection
	s.ClientsReadOnly[client.ID] = client
	s.clientsReadOnlyMutex.Unlock()

	initialWorldStateMessage, err := s.buildInitialWorldStateMessage()
	if err != nil {
		s.removeClientReadOnly(client)
		return
	}

	s.sendInitialWorldState(client, initialWorldStateMessage)

	// Defer deletion on disconnect
	defer s.removeClientReadOnly(client)

	// Initialise the ping loop (ensure client is connected) and write loop (for outbound messages)
	go client.PingLoop()
	// log.Println("client connected: ", client.ID)

	client.WriteLoop()
}

func (s *GServer) removeClientReadOnly(client *GServerClient) {
	s.clientsReadOnlyMutex.Lock()
	delete(s.ClientsReadOnly, client.ID)
	s.clientsReadOnlyMutex.Unlock()

	client.Close()
}

func (s *GServer) buildInitialWorldStateMessage() ([]byte, error) {
	// First message is world state
	serverInitialWorldStateMessage, err := messages.NewGServerInitialWorldStateMessage(s.WorldController.GameWorld)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial world state")
	}

	// Marshal the message
	return json.Marshal(serverInitialWorldStateMessage)
}

func (s *GServer) sendInitialWorldState(client *GServerClient, initialWorldStateMessage []byte) {
	// Send the inital message
	client.WriteMessage(initialWorldStateMessage)
}

func (s *GServer) clientCount() int {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	return len(s.Clients)
}

func (s *GServer) clientReadOnlyCount() int {
	s.clientsReadOnlyMutex.RLock()
	defer s.clientsReadOnlyMutex.RUnlock()

	return len(s.ClientsReadOnly)
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
		log.Printf(
			"TICK | slowest: %v, took: %v, target: %v, clients: %v, clients (ro): %v",
			peakTickSpeed,
			timeTaken,
			s.TickSpeed,
			s.clientCount(),
			s.clientReadOnlyCount(),
		)
	}
}

func (s *GServer) snapshotClients() map[string]*GServerClient {
	s.clientsMutex.RLock()
	currentClients := make(map[string]*GServerClient, len(s.Clients))
	maps.Copy(currentClients, s.Clients)
	s.clientsMutex.RUnlock()
	return currentClients
}

func (s *GServer) snapshotClientsReadOnly() map[string]*GServerClient {
	s.clientsReadOnlyMutex.RLock()
	currentClientsReadOnly := make(map[string]*GServerClient, len(s.ClientsReadOnly))
	maps.Copy(currentClientsReadOnly, s.ClientsReadOnly)
	s.clientsReadOnlyMutex.RUnlock()
	return currentClientsReadOnly
}

func (s *GServer) doTick() {
	s.tick++

	// Handle connections and disconnections first
	s.handleClientConnectionsAndDisconnections()

	// Get snapshot of current clients
	currentClients := s.snapshotClients()
	currentClientsReadOnly := s.snapshotClientsReadOnly()

	// Do the game world tick
	s.doGameWorldTick(currentClients)

	// Build diff from changes and send to clients
	diff := s.WorldController.BuildWorldDiff()
	err := s.MessageRouter.PushWorldDiffMessage(&diff)
	if err != nil {
		log.Printf("Failed to push world diff message: %s", err)
	}

	s.MessageRouter.Flush(currentClients, currentClientsReadOnly)
}

func (s *GServer) doGameWorldTick(currentClients map[string]*GServerClient) {
	// Do Client Actions
	for _, client := range currentClients {
		clientMessages := client.DrainMessages()
		s.handleMessages(client, clientMessages) // Client actions affect new game world
	}

	// Run tickers last
	s.WorldController.DoTickers()
	s.WorldController.DoNpcs()
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
			s.MessageRouter.PushGlobalLogMessage("GLOBAL", fmt.Sprintf("Player %s joined", clientID))
		}
	}

	// Handle queuedDisconnections
	for _, clientID := range disconnections {
		s.WorldController.DeletePlayer(clientID)
		s.MessageRouter.PushGlobalLogMessage("GLOBAL", fmt.Sprintf("Player %s disconnected", clientID))
	}
}

// @TODO - need a better way to handle this. Too much logic in the server
// Maybe some message processor - Cannot modify the world directly because i need to account for state?
// Or just process the messages - when collecting only one of each message type (the latest) can be used
func (s *GServer) handleMessages(
	client *GServerClient,
	msgs map[messages.GMessageType]*messages.GMessage,
) {
	for messageType, message := range msgs {
		switch messageType {

		case messages.TClientMoveMessage:
			moveData, err := messages.ParseGClientMoveMessageData(message.Data)
			if err != nil {
				log.Printf("error: failed to parse move data")
			}

			s.WorldController.MovePlayer(client, moveData.Dx, moveData.Dy)

		case messages.TClientInteractMessage:
			interactData, err := messages.ParseGClientInteractMessageData(message.Data)
			if err != nil {
				log.Printf("error: failed to parse interact data")
			}

			s.WorldController.InteractWith(client, interactData.InteractableInstanceID)
		case messages.TClientAttackNpcMessage:
			attackNpcData, err := messages.ParseGClientAttackNpcMessageData(message.Data)
			if err != nil {
				log.Printf("error: failed to parse attack npc data")
			}

			s.WorldController.AttackNpc(client, attackNpcData.NpcInstanceID)
		}

	}

}
