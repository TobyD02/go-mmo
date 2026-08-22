package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/google/uuid"
	"github.com/tobyd02/go-mmo/pkg/bbt_client"
	"github.com/tobyd02/go-mmo/pkg/client"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	clientID := uuid.NewString()
	gClient := client.NewGClient(false)
	worldState, err := gClient.Start(serverURI, clientID)

	if err != nil {
		log.Fatal(err)
	}

	defer gClient.StopAndCloseConnection()

	model := bbt_client.InitialModel(worldState, gClient)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}
