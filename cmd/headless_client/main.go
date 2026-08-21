package main

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/tobyd02/golang-mmo/pkg/client"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	clientID := uuid.NewString()

	c := client.NewGClient(true)

	log.Printf("Connecting to %s", serverURI)

	// Make it read only
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

	// client ticker

	ticker := time.NewTicker(client.GClientTickSpeed)
	defer ticker.Stop()

	for range ticker.C {
		c.Update() // Called first

		diff, err := c.ReadGameWorldDiff()
		if err != nil {
			log.Fatalf("Failed reading world diff: %v", err)
		}

		if diff != nil && !diff.IsEmpty() {
			log.Printf(
				"Received world diff: %s", time.Now(),
			)
		}
		// c.SendMoveMessage(rand.IntN(3)-1, rand.IntN(3)-1)
	}
}
