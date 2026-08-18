// Package game - global module for game logic
package game

import (
	"log"
	"maps"
	"reflect"
)

type GameWorldTile int

const (
	TileBlank GameWorldTile = iota
	TileEntity
)

type GameWorld struct {
	Map      [][]GameWorldTile
	Entities map[string]*GEntity
	Width    int
	Height   int
}

type GameWorldDiff struct {
	MapDiff      []GameWorldMapDiff  `json:"map_diff"`
	EntitiesDiff map[string]*GEntity `json:"entities_diff"`
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

	gameEntities := make(map[string]*GEntity)

	return &GameWorld{
		Map:      gameMap,
		Entities: gameEntities,
		Width:    width, Height: height,
	}
}

func (g *GameWorld) Clone() *GameWorld {
	clone := NewGameWorld(g.Width, g.Height)

	for id, entity := range g.Entities {
		entityCopy := *entity
		clone.Entities[id] = &entityCopy
	}

	for y := range g.Map {
		copy(clone.Map[y], g.Map[y])
	}
	return clone
}

func GenerateDiff(old *GameWorld, new *GameWorld) GameWorldDiff {
	diff := GameWorldDiff{
		EntitiesDiff: make(map[string]*GEntity),
	}

	for y, row := range old.Map {
		for x, tile := range row {
			if new.Map[y][x] != tile {
				diff.MapDiff = append(diff.MapDiff, GameWorldMapDiff{Pos: Vec2{X: x, Y: y}, Tile: new.Map[y][x]})
			}
		}
	}

	for id, oldEntity := range old.Entities {
		newEntity, exists := new.Entities[id]

		if exists && !reflect.DeepEqual(oldEntity, newEntity) {
			diff.EntitiesDiff[id] = newEntity
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

	if g.Entities == nil {
		g.Entities = make(map[string]*GEntity)
	}

	maps.Copy(g.Entities, diff.EntitiesDiff)
}

func (g *GameWorld) AddPlayer(playerID string, x, y int) {
	g.addEntity(&GEntity{
		ID: playerID,
		Pos: Vec2{
			X: x, Y: y,
		},
		Tags: &[]GTag{GPlayer},
	})
}

func (g *GameWorld) DeletePlayer(playerID string) {
	playerPos := g.Entities[playerID].Pos
	delete(g.Entities, playerID)

	g.Map[playerPos.Y][playerPos.X] = TileBlank
}

func (g *GameWorld) addEntity(entity *GEntity) {
	g.Map[entity.Pos.Y][entity.Pos.X] = TileEntity
	g.Entities[entity.ID] = entity
}

func (g *GameWorld) MoveEntity(entityID string, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	entity := g.Entities[entityID]

	if (entity.Pos.Y+dy < 0 || entity.Pos.Y+dy >= len(g.Map)) ||
		(entity.Pos.X+dx < 0 || entity.Pos.X+dx >= len(g.Map[entity.Pos.Y+dy])) {
		return
	}

	g.Map[entity.Pos.Y][entity.Pos.X] = TileBlank

	entity.Pos.X += dx
	entity.Pos.Y += dy

	g.Map[entity.Pos.Y][entity.Pos.X] = TileEntity

	log.Printf("WORLD | %s moved to x: %v y: %v", entity.ID, entity.Pos.X, entity.Pos.Y)
}
