package server

import (
	"fmt"
	"log"
	"math/rand/v2"

	"github.com/tobyd02/golang-mmo/pkg/game"
)

type GWorldController struct {
	GameWorld        *game.GameWorld
	getServerTick    func() int
	getMessageRouter func() *GMessageRouter

	changedPlayers       map[string]struct{}
	changedInteractables map[string]struct{}
	changedTiles         map[game.Vec2]struct{}
}

func NewGWorldController(worldWidth, worldHeight int, getTicker func() int, getMessageRouter func() *GMessageRouter) *GWorldController {
	return &GWorldController{
		GameWorld:        game.NewGameWorld(worldWidth, worldHeight),
		getServerTick:    getTicker,
		getMessageRouter: getMessageRouter,

		changedPlayers: make(map[string]struct{}),

		changedInteractables: make(map[string]struct{}),
		changedTiles:         make(map[game.Vec2]struct{}),
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

				if x != wc.GameWorld.SpawnPoint.X && y != wc.GameWorld.SpawnPoint.Y {
					if rand.IntN(100) == 1 {
						wc.AddInteractable(game.NewGInteractableInstance(x, y, "interactable.oak_tree"))
					}
				}
			}
		}
	}

}

func (wc *GWorldController) BuildWorldDiff() game.GameWorldDiff {
	diff := game.GameWorldDiff{
		PlayersDiff:       make(map[string]*game.GPlayer),
		InteractablesDiff: make(map[string]*game.GInteractableInstance),
	}

	for playerID := range wc.changedPlayers {
		player, exists := wc.GameWorld.Players[playerID]

		if exists {
			diff.PlayersDiff[playerID] = player
		} else {
			diff.PlayersDiff[playerID] = nil
		}
	}

	for id := range wc.changedInteractables {
		interactable, exists := wc.GameWorld.Interactables[id]

		if exists {
			diff.InteractablesDiff[id] = interactable
		} else {
			diff.InteractablesDiff[id] = nil
		}
	}

	for pos := range wc.changedTiles {
		diff.MapDiff = append(
			diff.MapDiff,
			game.GameWorldMapDiff{
				Pos:  pos,
				Tile: wc.GameWorld.Map[pos.Y][pos.X],
			},
		)
	}

	clear(wc.changedPlayers)
	clear(wc.changedInteractables)
	clear(wc.changedTiles)

	return diff
}

func (wc *GWorldController) AddInteractable(interactable *game.GInteractableInstance) {
	if wc.GameWorld.QueryInteractableAtPosition(interactable.Pos.X, interactable.Pos.Y) != nil {
		return // Only a single interactable in a tile
	}
	if wc.GameWorld.QueryMap(interactable.Pos.X, interactable.Pos.Y) != game.TileWalkable {
		return // Only on walkable tiles, not inside a wall
	}

	wc.GameWorld.Interactables[interactable.ID] = interactable
}

func (wc *GWorldController) SpawnNewPlayer(playerID string) error {
	return wc.AddPlayer(playerID, wc.GameWorld.SpawnPoint.X, wc.GameWorld.SpawnPoint.Y)
}

func (wc *GWorldController) AddPlayer(playerID string, x, y int) error {
	if wc.GameWorld.QueryMap(x, y) != game.TileWalkable {
		return fmt.Errorf("Cannot add player")
	}

	wc.addPlayer(game.NewGPlayer(playerID, x, y))

	wc.changedPlayers[playerID] = struct{}{}

	return nil
}

func (wc *GWorldController) DeletePlayer(playerID string) {
	if _, exists := wc.GameWorld.Players[playerID]; !exists {
		return
	}

	delete(wc.GameWorld.Players, playerID)

	wc.changedPlayers[playerID] = struct{}{}
}

func (wc *GWorldController) addPlayer(player *game.GPlayer) {
	wc.GameWorld.Players[player.ID] = player
}

func (wc *GWorldController) MovePlayer(client *GServerClient, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	player, exists := wc.GameWorld.Players[client.ID]
	if !exists { // Safe guard
		return
	}

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

	wc.changedPlayers[client.ID] = struct{}{}

	// wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("moved from %d to %d", player.Pos.X, player.Pos.Y))
	// log.Printf("WORLD | %s moved to x: %v y: %v", player.ID, player.Pos.X, player.Pos.Y)
}

func (wc *GWorldController) InteractWith(client *GServerClient, interactableInstanceID string) {
	player, exists := wc.GameWorld.Players[client.ID]
	if !exists { // Safe guard
		return
	}

	interactableInstance, exists := wc.GameWorld.Interactables[interactableInstanceID]
	if !exists { // Safe guard
		return
	}

	if !interactableInstance.PlayerCanOccupyOrWork(player) {
		return
	}

	interactableInstance.DoWork(wc.getServerTick())

	interactableName := game.GetInteractableNameFromRegistry(interactableInstance.InteractableID)
	wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("Worked %s", interactableName))

	if interactableInstance.WorkIsDone() {
		yields := interactableInstance.GetYieldAndTriggerCooldown()
		player.AddToInventory(yields)

		for itemID, yield := range yields {
			itemName := game.GetItemNameFromRegistry(itemID)
			wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprintf("Received %dx %s", yield, itemName))
		}

		if len(yields) == 0 {
			wc.getMessageRouter().PushClientLogMessage(client.ID, "CLIENT", fmt.Sprint("Received nothing"))
		}
	}

	wc.changedInteractables[interactableInstanceID] = struct{}{}
}

func (wc *GWorldController) DoTickers() {
	// @todo - need to tick occupancy as well.
	// - if occupant hasn't worked this tick, it should be cleared

	for _, interactable := range wc.GameWorld.Interactables {

		// If its occupied and hasn't been worked this tick
		if interactable.IsOccupied() && !interactable.DidWorkThisTick(wc.getServerTick()) {
			if interactable.OccupantCooldown <= 0 {
				interactable.ClearOccupant()
			} else {
				interactable.OccupantCooldown--
			}

			wc.changedInteractables[interactable.ID] = struct{}{}
		}

		if interactable.CurrentTickCooldown <= 0 {
			continue
		}

		interactable.CurrentTickCooldown--

		if interactable.CurrentTickCooldown <= 0 {
			interactable.OccupiedBy = ""
			interactable.CurrentTicksWorked = 0
			log.Printf("WORLD | %s interactable was reset", interactable.ID)
		}

		wc.changedInteractables[interactable.ID] = struct{}{}
	}
}
