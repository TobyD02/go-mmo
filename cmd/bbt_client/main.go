package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"github.com/tobyd02/golang-mmo/pkg/bbt_client"
	"github.com/tobyd02/golang-mmo/pkg/client"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	clientID := uuid.NewString()
	client := client.NewGClient(false)
	worldState, err := client.Start(serverURI, clientID)

	if err != nil {
		log.Fatal(err)
	}

	defer client.StopAndCloseConnection()

	model := bbt_client.InitialModel(worldState, client)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}
