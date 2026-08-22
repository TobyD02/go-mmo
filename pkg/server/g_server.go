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
	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/messages"
)

type InitialWorldStateResult struct {
	Payload []byte
	Err     error
}

type GServer struct {
	Clients      map[string]*GServerClient
	clientsMutex sync.RWMutex

	ClientsReadOnly      map[string]*GServerClient
	clientsReadOnlyMutex sync.RWMutex

	// owned by game loop. queued connections are read into these maps until they are ready to receive world updates
	pendingClients         map[string]*GServerClient
	pendingClientsReadOnly map[string]*GServerClient

	WorldController *GWorldController

	queuedConnections      []*GServerClient
	queuedConnectionsMutex sync.RWMutex

	queuedConnectionsReadOnly      []*GServerClient
	queuedConnectionsReadOnlyMutex sync.RWMutex

	queuedDisconnections []string

	initialWorldStateBuilding bool                         // is there currently a builder goroutine for the world state
	initialWorldStateResult   chan InitialWorldStateResult // channel for the builder goroutine to write to

	diffsSinceInitialWorldStateBuilding [][]byte

	upgrader websocket.Upgrader

	TickSpeed time.Duration
	tick      int

	MessageRouter *GMessageRouter
}

func NewGServer(tickSpeed time.Duration, worldWidth int, worldHeight int) *GServer {
	server := &GServer{
		Clients:         make(map[string]*GServerClient),
		ClientsReadOnly: make(map[string]*GServerClient),

		pendingClients:         make(map[string]*GServerClient),
		pendingClientsReadOnly: make(map[string]*GServerClient),

		queuedConnections:         make([]*GServerClient, 0),
		queuedConnectionsReadOnly: make([]*GServerClient, 0),

		queuedDisconnections: make([]string, 0),

		initialWorldStateResult: make(chan InitialWorldStateResult, 1),

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

	s.clientsMutex.RLock()

	if client.ID == "" {
		s.clientsMutex.RUnlock()

		log.Printf("client attempted to connect with empty ID")
		client.Close()
		return
	}

	if _, exists := s.Clients[client.ID]; exists {
		s.clientsMutex.RUnlock()

		log.Printf("client already connected: %s", client.ID)
		client.Close()
		return
	}
	s.clientsMutex.RUnlock()

	s.queuedConnectionsMutex.Lock()

	// Queue the connection

	s.queuedConnections = append(s.queuedConnections, client)
	s.queuedConnectionsMutex.Unlock()

	// Defer deletion on disconnect
	defer s.removeClient(client)

	// Initialise the ping loop (ensure client is connected) and write loop (for outbound messages)
	go client.WriteLoop()
	go client.PingLoop()
	// log.Println("client connected: ", client.ID)

	err = client.ReadLoop()
	if err != nil {
		fmt.Printf("client disconnected %s (%s)\n", client.ID, err)
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

	s.clientsReadOnlyMutex.RLock()

	if _, exists := s.ClientsReadOnly[client.ID]; exists {
		s.clientsReadOnlyMutex.RUnlock()

		log.Printf("(read only) client already connected: %s", client.ID)
		client.Close()
		return
	}

	s.clientsReadOnlyMutex.RUnlock()

	s.queuedConnectionsReadOnlyMutex.Lock()
	s.queuedConnectionsReadOnly = append(s.queuedConnectionsReadOnly, client)
	s.queuedConnectionsReadOnlyMutex.Unlock()

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
			"TICK | slowest: %-12v | took: %-12v | target: %-12v | clients: %-4v | clients (ro): %-4v",
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

	select {
	case initialState := <-s.initialWorldStateResult:
		s.initialWorldStateBuilding = false

		if initialState.Err != nil {
			log.Printf("intial world state errored %s\n", initialState.Err)
			for _, client := range s.pendingClients {
				client.Close()
			}

			for _, client := range s.pendingClientsReadOnly {
				client.Close()
			}

			clear(s.pendingClients)
			clear(s.pendingClientsReadOnly)
			s.diffsSinceInitialWorldStateBuilding = nil

			break
		}

		// send messages to clients
		for _, client := range s.pendingClients {
			// send initial world state
			s.sendInitialWorldState(client, initialState.Payload)

			// then send over collated diffs
			for _, diff := range s.diffsSinceInitialWorldStateBuilding {
				client.WriteMessage(diff)
			}

			s.clientsMutex.Lock()
			s.Clients[client.ID] = client
			s.clientsMutex.Unlock()

			delete(s.pendingClients, client.ID)
		}

		for _, client := range s.pendingClientsReadOnly {
			// send initial world state
			s.sendInitialWorldState(client, initialState.Payload)

			// then send over collated diffs
			for _, diff := range s.diffsSinceInitialWorldStateBuilding {
				client.WriteMessage(diff)
			}

			s.clientsReadOnlyMutex.Lock()
			s.ClientsReadOnly[client.ID] = client
			s.clientsReadOnlyMutex.Unlock()

			delete(s.pendingClientsReadOnly, client.ID)
		}

		s.diffsSinceInitialWorldStateBuilding = nil

	default:
		//pass for now
	}

	// Handle connections and disconnections first
	s.handleClientConnectionsAndDisconnections()

	// Get snapshot of current clients
	currentClients := s.snapshotClients()
	currentClientsReadOnly := s.snapshotClientsReadOnly()

	// Do the game world tick
	s.doGameWorldTick(currentClients)

	// Build diff from changes and send to clients
	diff := s.WorldController.BuildWorldDiff()
	payload, err := s.MessageRouter.PushWorldDiffMessage(&diff)
	if err != nil {
		log.Printf("Failed to push world diff message: %s", err)
	} else if s.initialWorldStateBuilding {
		s.diffsSinceInitialWorldStateBuilding = append(s.diffsSinceInitialWorldStateBuilding, payload)
	}

	s.MessageRouter.Flush(currentClients, currentClientsReadOnly)
}

func (s *GServer) doGameWorldTick(currentClients map[string]*GServerClient) {
	// @todo move to config/const
	simulateRangeX := config.ClientSimulationRangeX
	simulateRangeY := config.ClientSimulationRangeX

	npcSet := make(map[string]struct{})
	interactableSet := make(map[string]struct{})

	// Do Client Actions
	for _, client := range currentClients {
		clientMessages := client.DrainMessages()
		s.handleMessages(client, clientMessages) // Client actions affect new game world

		// Get surrounding NPCS into the set
		player := s.WorldController.GameWorld.Players[client.ID]
		if player == nil {
			continue
		}

		minX := player.Pos.X - simulateRangeX
		maxX := player.Pos.X + simulateRangeX
		minY := player.Pos.Y - simulateRangeY
		maxY := player.Pos.Y + simulateRangeY

		npcRangeSet := s.WorldController.npcSpatialIndex.QueryPosRange(minX, maxX, minY, maxY)
		interactableRangeSet := s.WorldController.interactableSpatialIndex.QueryPosRange(minX, maxX, minY, maxY)

		for npcID := range npcRangeSet {
			npcSet[npcID] = struct{}{}
		}

		for interactableID := range interactableRangeSet {
			interactableSet[interactableID] = struct{}{}
		}

	}

	log.Printf(
		"SIMUL | npcs: %d/%-12d | interactables: %d/%-12d\n",
		len(npcSet),
		len(s.WorldController.GameWorld.Npcs),
		len(interactableSet),
		len(s.WorldController.GameWorld.Interactables),
	)

	// Run tickers last
	s.WorldController.DoInteractables(interactableSet)
	s.WorldController.DoNpcs(npcSet)
}

func (s *GServer) getClient(clientID string) *GServerClient {
	s.clientsMutex.RLock()
	defer s.clientsMutex.RUnlock()

	return s.Clients[clientID]
}

func (s *GServer) handleClientConnectionsAndDisconnections() {
	s.queuedConnectionsMutex.Lock()
	connections := s.queuedConnections
	s.queuedConnections = nil
	s.queuedConnectionsMutex.Unlock()

	s.queuedConnectionsReadOnlyMutex.Lock()
	connectionsReadOnly := s.queuedConnectionsReadOnly
	s.queuedConnectionsReadOnly = nil
	s.queuedConnectionsReadOnlyMutex.Unlock()

	s.clientsMutex.Lock()
	disconnections := s.queuedDisconnections
	s.queuedDisconnections = nil
	s.clientsMutex.Unlock()

	if len(connections) > 0 {
		// Handle new connections
		for _, client := range connections {

			if err := s.WorldController.SpawnNewPlayer(client.ID); err != nil {
				client.Close()
				continue
			}

			s.pendingClients[client.ID] = client
		}
	}

	if len(connectionsReadOnly) > 0 {
		// Handle new connections
		for _, client := range connectionsReadOnly {
			s.pendingClientsReadOnly[client.ID] = client
		}
	}

	if len(s.pendingClients) > 0 || len(s.pendingClientsReadOnly) > 0 {
		s.startInitialWorldStateBuilding()
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

func (s *GServer) startInitialWorldStateBuilding() {
	if s.initialWorldStateBuilding {
		return
	}

	s.initialWorldStateBuilding = true

	initialWorldState := s.WorldController.GameWorld.Copy()

	go func(initialWorldState *game.GameWorld) {

		// First message is world state
		serverInitialWorldStateMessage, err := messages.NewGServerInitialWorldStateMessage(initialWorldState)
		if err != nil {
			s.initialWorldStateResult <- InitialWorldStateResult{
				Payload: nil,
				Err:     err,
			}
			return
		}

		// Marshal the message
		worldState, err := json.Marshal(serverInitialWorldStateMessage)
		if err != nil {
			log.Printf("failed to build initial world state: %s", err)
			s.initialWorldStateResult <- InitialWorldStateResult{
				Payload: nil,
				Err:     err,
			}
			return
		}

		s.initialWorldStateResult <- InitialWorldStateResult{
			Payload: worldState,
			Err:     nil,
		}
	}(initialWorldState)
}
