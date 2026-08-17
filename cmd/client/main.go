package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/client"
	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/server"
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

	err = conn.WriteJSON(server.GClientConnection{
		ID: uuid.NewString(),
	})
	if err != nil {
		log.Fatal(err)
	}

	world := game.NewGameWorld(50, 30)

	model := client.InitialModel(world, conn)

	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}
