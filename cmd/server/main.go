package main

import (
	"log"
	"net/http"
	"time"

	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/server"
)

func main() {

	// Get it to initialise it
	itemRegistry, err := game.GetItemRegistry()
	if err != nil {
		log.Fatalf("Failed to get item registry %s", err)
	}

	interactableRegistry, err := game.GetInteractableRegistry()
	if err != nil {
		log.Fatalf("Failed to get item registry %s", err)
	}

	log.Println("%v", itemRegistry)
	log.Println("%v", interactableRegistry)

	server := server.NewGServer(time.Millisecond*50, 200, 200)
	http.HandleFunc("/ws", server.HandleClientConnection)
	go server.GameLoop()

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
