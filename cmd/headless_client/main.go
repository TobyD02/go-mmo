package main

import (
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/tobyd02/go-mmo/pkg/client"
	"github.com/tobyd02/go-mmo/pkg/config"
)

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	clientID := uuid.NewString()

	// Make it read only
	c, err := client.NewGClient(true)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Connecting to %s", serverURI)

	err = c.Start(serverURI, clientID)
	if err != nil {
		log.Fatalf("Failed to start client: %v", err)
	}

	defer c.StopAndCloseConnection()

	log.Printf(
		"Connected! ClientID=%s",
		c.ClientID,
	)

	log.Println("Waiting for world diffs...")
	// client ticker
	ticker := time.NewTicker(config.ClientTickSpeed)
	defer ticker.Stop()

	for range ticker.C {
		err = c.Update() // Called first
		if err != nil {
			log.Fatalf("Failed to update client: %v", err)
		}

		diff, err := c.ReadGameWorldDiff()
		if err != nil {
			log.Fatalf("Failed reading world diff: %v", err)
		}

		if diff != nil && !diff.IsEmpty() {
			log.Printf(
				"Received world diff: %s", time.Now(),
			)
		}
		//_ = c.Move(rand.IntN(3)-1, rand.IntN(3)-1)
	}
}
