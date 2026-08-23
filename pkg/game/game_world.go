// Package game - game logic shared between server and client
package game

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/tobyd02/go-mmo/pkg/util"
)

type GameWorld struct {
	Map           [][]GameWorldTile
	Players       map[string]*GPlayer
	Npcs          map[string]*GNpcInstance
	Interactables map[string]*GInteractableInstance

	Width      int
	Height     int
	SpawnPoint util.Vec2
}

func NewGameWorld(worldFilePath string) (*GameWorld, error) {
	data, err := os.ReadFile(worldFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read world file path: %s", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	height := len(lines)
	var width int

	spawn := util.Vec2{X: 0, Y: 0}

	gameMap := make([][]GameWorldTile, 0, len(lines))
	for y, line := range lines {
		values := strings.Split(strings.TrimSpace(line), ",")
		row := make([]GameWorldTile, 0, len(values))

		if y == 0 {
			width = len(values)
		}

		for x, value := range values {
			n, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, err
			}

			tile := GameWorldTile(n)
			if tile == TileSpawn {
				spawn.X = x
				spawn.Y = y
			}
			row = append(row, GameWorldTile(n))
		}

		gameMap = append(gameMap, row)
	}

	log.Printf("%d, %d\n", spawn.X, spawn.Y)

	gamePlayers := make(map[string]*GPlayer)
	gameNpcs := make(map[string]*GNpcInstance)
	gameInteractables := make(map[string]*GInteractableInstance)

	log.Printf(
		"World dimensions: Width=%d Height=%d MapRows=%d FirstRowWidth=%d",
		width,
		height,
		len(gameMap),
		len(gameMap[0]),
	)

	return &GameWorld{
		Map:           gameMap,
		Players:       gamePlayers,
		Npcs:          gameNpcs,
		Interactables: gameInteractables,
		Width:         width, Height: height,
		SpawnPoint: spawn,
	}, nil
}

func (g *GameWorld) Copy() *GameWorld {
	if g == nil {
		return nil
	}

	// 1. Deep copy 2D slice
	mapCopy := make([][]GameWorldTile, len(g.Map))
	for y, row := range g.Map {
		rowCopy := make([]GameWorldTile, len(row))
		copy(rowCopy, row)
		mapCopy[y] = rowCopy
	}

	// 2. Deep copy Players map & instances
	playersCopy := make(map[string]*GPlayer, len(g.Players))
	for k, v := range g.Players {
		if v != nil {
			playerCopy := *v // Value copy of GPlayer struct
			playersCopy[k] = &playerCopy
		}
	}

	// 3. Deep copy Npcs map & instances
	npcsCopy := make(map[string]*GNpcInstance, len(g.Npcs))
	for k, v := range g.Npcs {
		if v != nil {
			npcCopy := *v // Value copy of GNpcInstance struct
			npcsCopy[k] = &npcCopy
		}
	}

	// 4. Deep copy Interactables map & instances
	interactablesCopy := make(map[string]*GInteractableInstance, len(g.Interactables))
	for k, v := range g.Interactables {
		if v != nil {
			interactableCopy := *v // Value copy of GInteractableInstance struct
			interactablesCopy[k] = &interactableCopy
		}
	}

	return &GameWorld{
		Map:           mapCopy,
		Players:       playersCopy,
		Npcs:          npcsCopy,
		Interactables: interactablesCopy,
		Width:         g.Width,
		Height:        g.Height,
		SpawnPoint:    g.SpawnPoint, // Copy by value
	}
}

func (g *GameWorld) ApplyDiff(diff *GameWorldDiff) {
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

	// Npcs
	if g.Npcs == nil {
		g.Npcs = make(map[string]*GNpcInstance)
	}

	for id, npc := range diff.NpcsDiff {
		if npc == nil {
			delete(g.Npcs, id)
		} else {
			g.Npcs[id] = npc
		}
	}

	// Interactables
	if g.Interactables == nil {
		g.Interactables = make(map[string]*GInteractableInstance)
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
	if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
		return TileWall
	}

	return g.Map[y][x]
}
