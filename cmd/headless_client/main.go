package main

import (
	"log"
	"math/rand/v2"
	"os"

	"github.com/google/uuid"
	"github.com/tobyd02/golang-mmo/pkg/client"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	clientID := uuid.NewString()

	c := client.NewGClient()

	log.Printf("Connecting to %s", serverURI)

	world, err := c.Start(serverURI, clientID)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	defer c.StopAndCloseConnection()

	log.Printf(
		"Connected! ClientID=%s Players=%d",
		c.ClientID,
		len(world.Players),
	)

	for id, player := range world.Players {
		log.Printf(
			"Player: id=%s position=(%d,%d)",
			id,
			player.Pos.X,
			player.Pos.Y,
		)
	}

	log.Println("Waiting for world diffs...")

	for {
		diff, err := c.ReadGameWorldDiff()
		if err != nil {
			log.Fatalf("Failed reading world diff: %v", err)
		}

		if !diff.IsEmpty() {
			log.Printf(
				"Received world diff: %+v",
				diff,
			)
		}

		c.SendMoveMessage(rand.IntN(3)-1, rand.IntN(3)-1)
	}
}
