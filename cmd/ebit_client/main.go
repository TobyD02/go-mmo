package main

import (
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tobyd02/go-mmo/pkg/client"
	"github.com/tobyd02/go-mmo/pkg/ebit_client"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	gClient, err := client.NewGClient(false)
	if err != nil {
		log.Fatal(err)
	}

	err = gClient.Start(serverURI, "toby")
	if err != nil {
		log.Fatal(err)
	}

	defer gClient.StopAndCloseConnection()

	gebitClient, err := ebit_client.NewGEbitClient(gClient)
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Map Editor")

	if err := ebiten.RunGame(gebitClient); err != nil {
		log.Fatal(err)
	}
}
