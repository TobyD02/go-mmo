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

const ServerTickSpeed = time.Millisecond * 20

type GServer struct {
	Clients              map[string]*GClient
	WorldController      *GWorldController
	clientsMutex         sync.RWMutex
	queuedConnections    []string
	queuedDisconnections []string
	upgrader             websocket.Upgrader
}

func NewGServer() *GServer {

	worldController := NewGWorldController(100, 20)
	worldController.SetupWorld(true)

	clients := make(map[string]*GClient)

	return &GServer{
		Clients:              clients,
		WorldController:      worldController,
		queuedConnections:    make([]string, 0),
		queuedDisconnections: make([]string, 0),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

}

func (s *GServer) HandleClientConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade to a websocket connection
	conn, err := s.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	defer conn.Close()

	client := NewGClient(conn)
	err = client.EstablishConnection()
	if err != nil {
		log.Printf("failed to establish client connection: %s", err)
		return
	}

	if _, exists := s.Clients[client.ID]; exists {
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("Client already connected"),
		)
		return
	}

	// Assign the client connection

	s.clientsMutex.Lock()
	s.Clients[client.ID] = client
	s.queuedConnections = append(s.queuedConnections, client.ID)
	s.clientsMutex.Unlock()

	// Defer deletion on disconnect
	defer s.removeClient(client)
	s.handleNewActiveClientConnection(client)

}

func (s *GServer) handleNewActiveClientConnection(client *GClient) {
	// First message is world state
	serverInitialWorldStateMessage, err := messages.NewGServerInitialWorldStateMessage(s.WorldController.GameWorld)
	if err != nil {
		fmt.Printf("Failed to get initial world state")
		return
	}

	// Marshal the message
	worldMsg, err := json.Marshal(serverInitialWorldStateMessage)
	if err != nil {
		fmt.Printf("Failed to get initial world state")
		return
	}

	// Send the inital message
	client.Conn.WriteMessage(websocket.TextMessage, worldMsg)

	// Initialise the ping loop (ensure client is connected)
	go client.PingLoop()
	log.Println("client connected: ", client.ID)

	err = client.ReadMessages()
	if err != nil {
		fmt.Printf("client disconnected %s", client.ID)
	}
}

func (s *GServer) removeClient(client *GClient) {
	s.clientsMutex.Lock()

	delete(s.Clients, client.ID)
	s.queuedDisconnections = append(s.queuedDisconnections, client.ID)
	s.clientsMutex.Unlock()

	client.Conn.Close()
}

func (s *GServer) GameLoop() {
	ticker := time.NewTicker(ServerTickSpeed)
	defer ticker.Stop()

	for range ticker.C {
		s.doTick()
	}
}

func (s *GServer) doTick() {
	// Get copy of current clients using Mutex
	s.clientsMutex.RLock()
	currentClients := make(map[string]*GClient, len(s.Clients))
	maps.Copy(currentClients, s.Clients)
	s.clientsMutex.RUnlock()

	// Get copy of current game world state
	oldState := s.WorldController.CloneWorld()

	// Do the game world tick
	s.doGameWorldTick(currentClients)

	// Generate diff to old state to send to clients
	diff := s.WorldController.GenerateWorldDiff(oldState)
	msg, err := json.Marshal(diff)

	if err != nil {
		fmt.Printf("Failed to do tick: %s", err)
		return
	}

	// Relay updates to clients
	for _, client := range s.Clients {
		client.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}

func (s *GServer) doGameWorldTick(currentClients map[string]*GClient) {
	s.handleClientConnectionsAndDisconnections()

	// Do Client Actions
	for clientID, client := range currentClients {
		clientMessages := client.DrainMessages()
		s.handleMessages(clientID, clientMessages) // Client actions affect new game world
	}

	// Do tickers before generating diff
	s.WorldController.DoTickers()
}

func (s *GServer) handleClientConnectionsAndDisconnections() {
	// Handle new connections
	for _, clientID := range s.queuedConnections {
		err := s.WorldController.AddPlayer(clientID, 10, 5)
		if err != nil {
			s.removeClient(s.Clients[clientID])
		}
	}

	s.queuedConnections = s.queuedConnections[:0]

	// Handle queuedDisconnections
	for _, clientID := range s.queuedDisconnections {
		s.WorldController.DeletePlayer(clientID)
	}

	s.queuedDisconnections = s.queuedDisconnections[:0]
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
