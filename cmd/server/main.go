package main

import (
	"log"
	"net/http"
	"time"

	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/server"
)

func main() {
	// Initialise Registries
	_, err := game.GetItemRegistry()
	if err != nil {
		log.Fatalf("Failed to get item registry %s", err)
	}

	_, err = game.GetInteractableRegistry()
	if err != nil {
		log.Fatalf("Failed to get interactable registry %s", err)
	}

	npcRegistry, err := game.GetNpcRegistry()
	if err != nil {
		log.Fatalf("Failed to get npc registry %s", err)
	}

	log.Println("Npc Registry: %v", npcRegistry)

	gServer := server.NewGServer(time.Millisecond*200, 5000, 5000)
	http.HandleFunc("/ws", gServer.HandleClientConnection)
	http.HandleFunc("/ws/ro", gServer.HandleClientConnectionReadOnly) // Read only websocket
	go gServer.GameLoop()

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
