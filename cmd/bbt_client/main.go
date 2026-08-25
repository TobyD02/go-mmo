package main

import (
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/tobyd02/go-mmo/pkg/bbt_client"
	"github.com/tobyd02/go-mmo/pkg/client"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	//clientID := uuid.NewString()

	gClient, err := client.NewGClient(false)
	if err != nil {
		log.Fatal(err)
	}

	err = gClient.Start(serverURI, "toby")
	if err != nil {
		log.Fatal(err)
	}

	defer gClient.StopAndCloseConnection()

	model := bbt_client.InitialModel(gClient)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		log.Fatal(err)
	}
}
