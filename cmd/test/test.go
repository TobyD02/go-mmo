package main

import (
	"log"
	"math/rand/v2"
	"os"

	"github.com/tobyd02/go-mmo/pkg/database"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/server"
	"github.com/tobyd02/go-mmo/pkg/util"
)

type TestRow struct {
	id   int
	name string
}

func main() {
	os.Setenv("SQLITE_PATH", "./test.db")
	db := database.NewDBSqlite()
	serverDB := server.NewGServerDB(db)

	err := serverDB.Init()
	if err != nil {
		log.Printf("%s", err)
		return
	}

	// try to load player first
	playerID := "test_id"
	player, err := serverDB.LoadPlayer(playerID)
	if err != nil {
		log.Printf("Failed to load player, creating new one: %s", err)
		player = game.NewGPlayer(playerID, 1, 0)
	}

	// Get Item Registry, so we can add a bunch of random items
	itemRegistry, err := game.GetItemRegistry()
	if err != nil {
		log.Printf("%s", err)
		return
	}

	for range 10 {
		itemName, err := util.GetRandomIDFromRegistry(itemRegistry)
		if err != nil {
			continue
		}
		itemAmount := rand.IntN(10)

		player.AddToInventory(map[string]int{itemName: itemAmount})
	}

	// Now that we've added a bunch of random items, save the player

	err = serverDB.SavePlayer(player)
	if err != nil {
		log.Printf("%s", err)
		return
	}

	// Finally, load the player again, and print out stats
	player, err = serverDB.LoadPlayer("test_id")
	if err != nil {
		log.Printf("%s", err)
		return
	}

	log.Printf("ID %-8s | X %-4d | Y %-4d\n", player.ID, player.Pos.X, player.Pos.Y)
	log.Println("Inventory:")
	for itemName, amount := range player.Inventory {
		log.Printf("\t%s: %d", itemName, amount)
	}

}
