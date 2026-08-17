package main

import (
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/server"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type PlayerPos struct {
	X int `json:"x"`
	Y int `json:"y"`
}

var clients = make(map[string]*server.GClient)
var worldState = make(map[string]*PlayerPos)

var clientsMutex sync.RWMutex

func main() {
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
	worldState[client.ID] = &PlayerPos{25, 15}

	clientsMutex.Unlock()

	go client.PingLoop()

	// Defer deletion on disconnect
	defer removeClient(client)

	defer delete(clients, client.ID)
	log.Println("client connected: ", client.ID)

	err = client.ReadActions()
	if err != nil {
		fmt.Printf("client disconnected %s", client.ID)
	}

}

func removeClient(client *server.GClient) {
	delete(clients, client.ID)
	client.Conn.Close()
	delete(worldState, client.ID)
}

func gameLoop() {

	ticker := time.NewTicker(1000 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		doTick()
	}
}

func doTick() {
	fmt.Println("doing tick")

	// updates := make(map[string]any)

	// world is 50 x 30

	clientsMutex.RLock()

	currentClients := make(map[string]*server.GClient, len(clients))
	maps.Copy(currentClients, clients)

	clientsMutex.RUnlock()

	for clientID, client := range currentClients {
		clientActions := client.DrainActions()

		dx, dy := 0, 0

		for _, a := range clientActions {
			dx = a.Dx
			dy = a.Dy
		}

		ws := worldState[clientID]
		ws.X = ws.X + dx
		ws.Y = ws.Y + dy
		// updates[clientID] = clientActions
	}

	msg, err := json.Marshal(worldState)

	if err != nil {
		fmt.Printf("Failed to do tick: %s", err)
		return
	}

	for _, client := range clients {
		client.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}
