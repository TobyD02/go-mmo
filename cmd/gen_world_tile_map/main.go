package main

import (
	"log"

	"github.com/tobyd02/go-mmo/pkg/game"
)

func main() {

	gameWorld, err := game.LoadGameWorld("./data/world_new.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println(gameWorld)

}
