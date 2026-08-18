package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/client"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	conn, _, err := websocket.DefaultDialer.Dial(
		serverURI+"/ws",
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// Connect to server
	connected, err := messages.NewGClientConnectedMessage(uuid.NewString())
	if err != nil {
		log.Fatal(err)
	}

	err = conn.WriteJSON(connected)

	if err != nil {
		log.Fatal(err)
	}

	// Wait for initial world state
	var msg messages.GMessage
	if err := conn.ReadJSON(&msg); err != nil {
		log.Fatalf("Expected world state, got message type %v", msg.Type)
	}

	parsedData, err := messages.ParseGServerInitialWorldStateData(msg.Data)
	if err != nil {
		log.Fatalf("Failed to parse initial world state message")
	}

	world := parsedData.InitialWorldState

	model := client.InitialModel(world, conn)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}
