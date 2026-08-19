package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/game_common"
	"github.com/tobyd02/golang-mmo/pkg/messages"
	"github.com/tobyd02/golang-mmo/pkg/server"
)

const TickSpeed = time.Millisecond * 200

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[string]*server.GClient)
var clientsMutex sync.RWMutex
var gameWorld *game_common.GameWorld

var queuedConnections = make([]string, 0)
var queuedDisconnections = make([]string, 0)

func main() {
	gameWorld = game_common.NewGameWorld(100, 20)

	for y, row := range gameWorld.Map {
		for x := range row {
			if gameWorld.Map[y][x] != game_common.TileWalkable {
				continue
			}

			if rand.IntN(20) == 1 {
				gameWorld.AddInteractable(game_common.NewGInteractable(x, y, game_common.TestItem))
			}
		}
	}

	http.HandleFunc("/ws", clientConnection)

	go gameLoop()

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}

func clientConnection(w http.ResponseWriter, r *http.Request) {
	// Upgrade to a websocket connection
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("upgrade:", err)
		return
	}

	defer conn.Close()

	client := server.NewGClient(conn)
	err = client.EstablishConnection()
	if err != nil {
		log.Printf("failed to establish client connection: %s", err)
		return
	}

	if _, exists := clients[client.ID]; exists {
		conn.WriteMessage(
			websocket.TextMessage,
			[]byte("Client already connected"),
		)
		return
	}

	clientsMutex.Lock()
	// Assig the client connection
	clients[client.ID] = client
	queuedConnections = append(queuedConnections, client.ID)
	// gameWorldDiff[client.ID] = &PlayerPos{25, 15}

	clientsMutex.Unlock()

	// Defer deletion on disconnect
	defer removeClient(client)

	// First message is world state
	serverInitialWorldStateMessage, err := messages.NewGServerInitialWorldStateMessage(gameWorld)
	if err != nil {
		fmt.Printf("Failed to get initial world state")
		return
	}

	worldMsg, err := json.Marshal(serverInitialWorldStateMessage)
	if err != nil {
		fmt.Printf("Failed to get initial world state")
		return
	}

	fmt.Printf("%v", gameWorld.Interactables)

	client.Conn.WriteMessage(websocket.TextMessage, worldMsg)

	go client.PingLoop()
	log.Println("client connected: ", client.ID)

	err = client.ReadMessages()
	if err != nil {
		fmt.Printf("client disconnected %s", client.ID)
	}

}

func removeClient(client *server.GClient) {
	clientsMutex.Lock()

	delete(clients, client.ID)
	queuedDisconnections = append(queuedDisconnections, client.ID)
	clientsMutex.Unlock()

	client.Conn.Close()
}

func gameLoop() {

	ticker := time.NewTicker(TickSpeed)
	defer ticker.Stop()

	for range ticker.C {
		doTick()
	}
}

func doTick() {
	// fmt.Println("doing tick")

	clientsMutex.RLock()

	currentClients := make(map[string]*server.GClient, len(clients))
	maps.Copy(currentClients, clients)
	clientsMutex.RUnlock()

	newGameWorld := gameWorld.Clone() // Copy gameWorld struct - it may be a pointer or contain pointers - they should be copied also

	// Handle new connections
	for _, clientID := range queuedConnections {
		err := newGameWorld.AddPlayer(clientID, 10, 5)
		if err != nil {
			removeClient(clients[clientID])
		}
	}

	queuedConnections = queuedConnections[:0]

	// Handle queuedDisconnections
	for _, clientID := range queuedDisconnections {
		newGameWorld.DeletePlayer(clientID)
	}

	queuedDisconnections = queuedDisconnections[:0]

	// Do Client Actions

	for clientID, client := range currentClients {
		clientMessages := client.DrainMessages()

		handleMessages(clientID, clientMessages, newGameWorld) // Client actions affect new game world
	}

	// Do tickers before generating diff
	newGameWorld.DoTickers()

	// Generate diff that can be sent to clients
	diff := game_common.GenerateDiff(gameWorld, newGameWorld)
	msg, err := json.Marshal(diff)

	// -------------------------

	if err != nil {
		fmt.Printf("Failed to do tick: %s", err)
		return
	}

	for _, client := range clients {
		client.Conn.WriteMessage(websocket.TextMessage, msg)
	}

	// Set curren t game world to new game world
	gameWorld = newGameWorld
}

func handleMessages(clientID string, msgs []*messages.GMessage, newGameWorld *game_common.GameWorld) {
	dx, dy := 0, 0

	// if len(msgs) > 0 {
	// 	fmt.Printf("Handling Messages: %v\n", msgs)
	// }

	for _, m := range msgs {
		if m.Type == messages.TClientMoveMessage {
			moveData, err := messages.ParseGClientMoveMessageData(m.Data)
			if err != nil {
				log.Printf("error: failed to parse move data")
			}

			// Assign them to only the latest received movement message - 1 tile per tick
			dx = moveData.Dx
			dy = moveData.Dy
		}

		if m.Type == messages.TClientInteractMessage {
			interactData, err := messages.ParseGClientInteractMessageData(m.Data)
			if err != nil {
				log.Printf("error: failed to parse move data")
			}

			newGameWorld.InteractWith(clientID, interactData.InteractableID)
		}
	}

	newGameWorld.MovePlayer(clientID, dx, dy)
}
