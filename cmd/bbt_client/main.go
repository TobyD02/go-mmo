package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/tobyd02/go-mmo/pkg/bbt_client"
	"github.com/tobyd02/go-mmo/pkg/client"
	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	//clientID := uuid.NewString()

	gClient := client.NewGClient(false)
	initialWorldState, err := game.NewGameWorld(config.GameWorldFilePath)

	if err != nil {
		log.Fatalf("Failed to load initial world state: %s", err)
	}

	worldState, err := gClient.Start(serverURI, "toby", initialWorldState)

	if err != nil {
		log.Fatal(err)
	}

	defer gClient.StopAndCloseConnection()

	model := bbt_client.InitialModel(worldState, gClient)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}
