package server

import (
	"fmt"
	"log"
	"math/rand/v2"
	"reflect"

	"github.com/tobyd02/golang-mmo/pkg/game"
)

type GWorldController struct {
	GameWorld *game.GameWorld
}

func NewGWorldController(worldWidth, worldHeight int) *GWorldController {
	return &GWorldController{
		GameWorld: game.NewGameWorld(worldWidth, worldHeight),
	}
}

func (wc *GWorldController) SetupWorld(
	edgeWalls bool,
) {
	for y, row := range wc.GameWorld.Map {
		for x := range row {
			if (x == 0 || x == wc.GameWorld.Width-1) || (y == 0 || y == wc.GameWorld.Height-1) {
				if edgeWalls {
					wc.GameWorld.Map[y][x] = game.TileWall
				}
			} else {
				// Not on a wall, so spawn an interactable (sometimes)
				if rand.IntN(100) == 1 {
					wc.AddInteractable(game.NewGInteractable(x, y, game.TestItem))
				}
			}
		}
	}

}

func (wc *GWorldController) CloneWorld() *game.GameWorld {
	clone := game.NewGameWorld(wc.GameWorld.Width, wc.GameWorld.Height)

	for id, player := range wc.GameWorld.Players {
		playerCopy := *player
		clone.Players[id] = &playerCopy
	}

	for id, interactable := range wc.GameWorld.Interactables {
		interactableCopy := *interactable
		clone.Interactables[id] = &interactableCopy
	}

	for y := range wc.GameWorld.Map {
		copy(clone.Map[y], wc.GameWorld.Map[y])
	}
	return clone
}

// GenerateWorldDiff Generates diff from and older world state
func (wc *GWorldController) GenerateWorldDiff(old *game.GameWorld) game.GameWorldDiff {
	diff := game.GameWorldDiff{
		PlayersDiff:       make(map[string]*game.GPlayer),
		InteractablesDiff: make(map[string]*game.GInteractable),
	}

	for y, row := range old.Map {
		for x, tile := range row {
			if wc.GameWorld.Map[y][x] != tile {
				diff.MapDiff = append(diff.MapDiff, game.GameWorldMapDiff{Pos: game.Vec2{X: x, Y: y}, Tile: wc.GameWorld.Map[y][x]})
			}
		}
	}

	// Added / changed players
	for id, newPlayer := range wc.GameWorld.Players {
		oldPlayer, exists := old.Players[id]

		if !exists || !reflect.DeepEqual(oldPlayer, newPlayer) {
			diff.PlayersDiff[id] = newPlayer
		}
	}

	// Deleted Players
	for id := range old.Players {
		if _, exists := wc.GameWorld.Players[id]; !exists {
			diff.PlayersDiff[id] = nil
		}
	}

	// Added / changed interactables
	for id, newInteractable := range wc.GameWorld.Interactables {
		oldInteractable, exists := old.Interactables[id]

		if !exists || !reflect.DeepEqual(oldInteractable, newInteractable) {
			diff.InteractablesDiff[id] = newInteractable
		}
	}

	// Deleted Interactables
	for id := range old.Interactables {
		if _, exists := wc.GameWorld.Interactables[id]; !exists {
			diff.InteractablesDiff[id] = nil
		}
	}

	return diff
}

func (wc *GWorldController) AddInteractable(interactable *game.GInteractable) {
	if wc.GameWorld.QueryInteractableAtPosition(interactable.Pos.X, interactable.Pos.Y) != nil {
		return // Only a single interactable in a tile
	}
	if wc.GameWorld.QueryMap(interactable.Pos.X, interactable.Pos.Y) != game.TileWalkable {
		return // Only on walkable tiles, not inside a wall
	}

	wc.GameWorld.Interactables[interactable.ID] = interactable
}

func (wc *GWorldController) AddPlayer(playerID string, x, y int) error {
	if wc.GameWorld.QueryMap(x, y) != game.TileWalkable {
		return fmt.Errorf("Cannot add player")
	}

	wc.addPlayer(&game.GPlayer{
		ID: playerID,
		Pos: game.Vec2{
			X: x, Y: y,
		},
	})

	return nil
}

func (wc *GWorldController) DeletePlayer(playerID string) {
	delete(wc.GameWorld.Players, playerID)
}

func (wc *GWorldController) addPlayer(player *game.GPlayer) {
	wc.GameWorld.Players[player.ID] = player
}

func (wc *GWorldController) MovePlayer(playerID string, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	player := wc.GameWorld.Players[playerID]

	newX := player.Pos.X + dx
	newY := player.Pos.Y + dy

	if (player.Pos.Y+dy < 0 || player.Pos.Y+dy >= len(wc.GameWorld.Map)) ||
		(player.Pos.X+dx < 0 || player.Pos.X+dx >= len(wc.GameWorld.Map[player.Pos.Y+dy])) {
		return
	}

	if wc.GameWorld.QueryMap(newX, newY) != game.TileWalkable {
		return // Cannot move to unwalkable tile
	}

	if wc.GameWorld.QueryInteractableAtPosition(newX, newY) != nil {
		return // Cannot move over interactables?
	}

	player.Pos.X += dx
	player.Pos.Y += dy

	log.Printf("WORLD | %s moved to x: %v y: %v", player.ID, player.Pos.X, player.Pos.Y)
}

func (wc *GWorldController) InteractWith(playerId string, interactableId string) {
	player := wc.GameWorld.Players[playerId]
	interactable := wc.GameWorld.Interactables[interactableId]

	if player.Pos.Distance(interactable.Pos) > 1 {
		return // Cannot
	}

	if interactable.CurrentTickCooldown != 0 {
		return // Interactable is on cooldown
	}

	if interactable.OccupiedBy == "" {
		interactable.OccupiedBy = playerId
	} else if interactable.OccupiedBy != playerId {
		return // Cannot - occupied by someone else
	}

	log.Printf("WORLD | %s interacted with %s", playerId, interactableId)

	interactable.CurrentTicksWorked++
	if interactable.CurrentTicksWorked >= interactable.TickWorkForYield {
		interactable.CurrentTickCooldown = interactable.MaxTickCooldown

		yieldAmount := rand.IntN(interactable.YieldAmountMax-interactable.YieldAmountMin) + interactable.YieldAmountMin

		player.AddToInventory(&interactable.Yield, yieldAmount)
		// log.Printf("WORLD | %s %v yielded %s from %s", playerId, yieldAmount, interactable.Yield.Name, interactableId)

		log.Printf("WORLD | Player %s inventory: %v", playerId, player.Inventory)

		for _, i := range player.Inventory {
			log.Printf("\t ITEM | %s | %v", i.Item.Name, i.Quantity)
		}
	}
}

func (wc *GWorldController) DoTickers() {
	for _, interactable := range wc.GameWorld.Interactables {
		if interactable.CurrentTickCooldown <= 0 {
			continue
		}

		interactable.CurrentTickCooldown--

		if interactable.CurrentTickCooldown <= 0 {
			interactable.OccupiedBy = ""
			interactable.CurrentTicksWorked = 0
			log.Printf("WORLD | %s interactable was reset", interactable.ID)
		}
	}
}
