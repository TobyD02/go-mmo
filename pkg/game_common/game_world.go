// Package game - global module for game logic
package game_common

import (
	"fmt"
	"log"
	"math/rand/v2"
	"reflect"
)

type GameWorld struct {
	Map           [][]GameWorldTile
	Players       map[string]*GPlayer
	Interactables map[string]*GInteractable
	Width         int
	Height        int
}

type GameWorldDiff struct {
	MapDiff           []GameWorldMapDiff        `json:"map_diff"`
	PlayersDiff       map[string]*GPlayer       `json:"players_diff"`
	InteractablesDiff map[string]*GInteractable `json:"interactables_diff"`
}

type GameWorldMapDiff struct {
	Pos  Vec2          `json:"position"`
	Tile GameWorldTile `json:"tile"`
}

func NewGameWorld(width, height int) *GameWorld {
	gameMap := make([][]GameWorldTile, height)
	for y := range gameMap {
		gameMap[y] = make([]GameWorldTile, width)
	}

	// Set walls
	for y, row := range gameMap {
		for x := range row {
			if (x == 0 || x == width-1) || (y == 0 || y == height-1) {
				gameMap[y][x] = TileWall
			}
		}
	}

	gamePlayers := make(map[string]*GPlayer)
	gameInteractables := make(map[string]*GInteractable)

	return &GameWorld{
		Map:           gameMap,
		Players:       gamePlayers,
		Interactables: gameInteractables,
		Width:         width, Height: height,
	}
}

func (g *GameWorld) Clone() *GameWorld {
	clone := NewGameWorld(g.Width, g.Height)

	for id, player := range g.Players {
		playerCopy := *player
		clone.Players[id] = &playerCopy
	}

	for id, interactable := range g.Interactables {
		interactableCopy := *interactable
		clone.Interactables[id] = &interactableCopy
	}

	for y := range g.Map {
		copy(clone.Map[y], g.Map[y])
	}
	return clone
}

func GenerateDiff(old *GameWorld, new *GameWorld) GameWorldDiff {
	diff := GameWorldDiff{
		PlayersDiff:       make(map[string]*GPlayer),
		InteractablesDiff: make(map[string]*GInteractable),
	}

	for y, row := range old.Map {
		for x, tile := range row {
			if new.Map[y][x] != tile {
				diff.MapDiff = append(diff.MapDiff, GameWorldMapDiff{Pos: Vec2{X: x, Y: y}, Tile: new.Map[y][x]})
			}
		}
	}

	// Added / changed players
	for id, newPlayer := range new.Players {
		oldPlayer, exists := old.Players[id]

		if !exists || !reflect.DeepEqual(oldPlayer, newPlayer) {
			diff.PlayersDiff[id] = newPlayer
		}
	}

	// Deleted Players
	for id := range old.Players {
		if _, exists := new.Players[id]; !exists {
			diff.PlayersDiff[id] = nil
		}
	}

	// Added / changed interactables
	for id, newInteractable := range new.Interactables {
		oldInteractable, exists := old.Interactables[id]

		if !exists || !reflect.DeepEqual(oldInteractable, newInteractable) {
			diff.InteractablesDiff[id] = newInteractable
		}
	}

	// Deleted Interactables
	for id := range old.Interactables {
		if _, exists := new.Interactables[id]; !exists {
			diff.InteractablesDiff[id] = nil
		}
	}

	return diff
}

func (g *GameWorld) ApplyDiff(diff GameWorldDiff) {
	for _, mapDiff := range diff.MapDiff {
		pos := mapDiff.Pos
		tile := mapDiff.Tile

		g.Map[pos.Y][pos.X] = tile
	}

	// Players
	if g.Players == nil {
		g.Players = make(map[string]*GPlayer)
	}

	for id, player := range diff.PlayersDiff {
		if player == nil {
			delete(g.Players, id)
		} else {
			g.Players[id] = player
		}
	}

	// Interactables
	if g.Interactables == nil {
		g.Interactables = make(map[string]*GInteractable)
	}

	for id, interactable := range diff.InteractablesDiff {
		if interactable == nil {
			delete(g.Interactables, id)
		} else {
			g.Interactables[id] = interactable
		}
	}
}

func (g *GameWorld) QueryMap(x, y int) GameWorldTile {
	if x <= 0 || x >= g.Width || y <= 0 || y >= g.Height {
		return TileWall
	}

	return g.Map[y][x]
}

func (g *GameWorld) QueryPlayersAtPosition(x, y int) map[string]*GPlayer {
	players := make(map[string]*GPlayer, 0)
	for playerID, e := range g.Players {
		if e.Pos.X == x && e.Pos.Y == y {
			players[playerID] = e
		}
	}

	return players
}

func (g *GameWorld) QueryInteractableAtPosition(x, y int) *GInteractable { // singular since there can only be one
	for _, i := range g.Interactables {
		if i.Pos.X == x && i.Pos.Y == y {
			return i
		}
	}

	return nil
}

func (g *GameWorld) AddInteractable(interactable *GInteractable) {
	if g.QueryInteractableAtPosition(interactable.Pos.X, interactable.Pos.Y) != nil {
		return // Only a single interactable in a tile
	}
	if g.QueryMap(interactable.Pos.X, interactable.Pos.Y) != TileWalkable {
		return // Only on walkable tiles, not inside a wall
	}

	g.Interactables[interactable.ID] = interactable
}

func (g *GameWorld) AddPlayer(playerID string, x, y int) error {
	if g.QueryMap(x, y) != TileWalkable {
		return fmt.Errorf("Cannot add player")
	}

	g.addPlayer(&GPlayer{
		ID: playerID,
		Pos: Vec2{
			X: x, Y: y,
		},
	})

	return nil
}

func (g *GameWorld) DeletePlayer(playerID string) {
	delete(g.Players, playerID)
}

func (g *GameWorld) addPlayer(player *GPlayer) {
	g.Players[player.ID] = player
}

func (g *GameWorld) MovePlayer(playerID string, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	player := g.Players[playerID]

	newX := player.Pos.X + dx
	newY := player.Pos.Y + dy

	if (player.Pos.Y+dy < 0 || player.Pos.Y+dy >= len(g.Map)) ||
		(player.Pos.X+dx < 0 || player.Pos.X+dx >= len(g.Map[player.Pos.Y+dy])) {
		return
	}

	if g.QueryMap(newX, newY) != TileWalkable {
		return // Cannot move to unwalkable tile
	}

	if g.QueryInteractableAtPosition(newX, newY) != nil {
		return // Cannot move over interactables?
	}

	player.Pos.X += dx
	player.Pos.Y += dy

	log.Printf("WORLD | %s moved to x: %v y: %v", player.ID, player.Pos.X, player.Pos.Y)
}

func (g *GameWorld) InteractWith(playerId string, interactableId string) {
	player := g.Players[playerId]
	interactable := g.Interactables[interactableId]

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

func (g *GameWorld) DoTickers() {
	for _, interactable := range g.Interactables {
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
