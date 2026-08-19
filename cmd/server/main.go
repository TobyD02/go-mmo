package main

import (
	"log"
	"net/http"

	"github.com/tobyd02/golang-mmo/pkg/server"
)

func main() {

	server := server.NewGServer()
	http.HandleFunc("/ws", server.HandleClientConnection)
	go server.GameLoop()

	log.Println("server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
