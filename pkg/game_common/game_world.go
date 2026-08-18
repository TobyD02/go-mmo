// Package game - global module for game logic
package game

import (
	"fmt"
	"log"
	"reflect"
)

type GameWorldTile int

const (
	TileWalkable GameWorldTile = iota
	TileWall
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

	// Set walls
	for y, row := range gameMap {
		for x := range row {
			if (x == 0 || x == width-1) || (y == 0 || y == height-1) {
				gameMap[y][x] = TileWall
			}
		}
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

	// Added / changed entities
	for id, newEntity := range new.Entities {
		oldEntity, exists := old.Entities[id]

		if !exists || !reflect.DeepEqual(oldEntity, newEntity) {
			diff.EntitiesDiff[id] = newEntity
		}
	}

	// Deleted Entities
	for id := range old.Entities {
		if _, exists := new.Entities[id]; !exists {
			diff.EntitiesDiff[id] = nil
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

	for id, entity := range diff.EntitiesDiff {
		if entity == nil {
			delete(g.Entities, id)
		} else {
			g.Entities[id] = entity
		}
	}
}

func (g *GameWorld) QueryMap(x, y int) GameWorldTile {
	if x <= 0 || x >= g.Width || y <= 0 || y >= g.Height {
		return TileWall
	}

	return g.Map[y][x]
}

func (g *GameWorld) QueryEntitiesAtPosition(x, y int) map[string]*GEntity {
	entities := make(map[string]*GEntity, 0)
	for entityID, e := range g.Entities {
		if e.Pos.X == x && e.Pos.Y == y {
			entities[entityID] = e
		}
	}

	return entities
}

func (g *GameWorld) AddPlayer(playerID string, x, y int) error {
	if g.QueryMap(x, y) != TileWalkable {
		return fmt.Errorf("Cannot add player")
	}

	g.addEntity(&GEntity{
		ID: playerID,
		Pos: Vec2{
			X: x, Y: y,
		},
		Tags: &[]GTag{GPlayer},
	})

	return nil
}

func (g *GameWorld) DeletePlayer(playerID string) {
	delete(g.Entities, playerID)
}

func (g *GameWorld) addEntity(entity *GEntity) {
	g.Entities[entity.ID] = entity
}

func (g *GameWorld) MoveEntity(entityID string, dx, dy int) {
	if dx == 0 && dy == 0 {
		return
	}

	entity := g.Entities[entityID]

	newX := entity.Pos.X + dx
	newY := entity.Pos.Y + dy

	if (entity.Pos.Y+dy < 0 || entity.Pos.Y+dy >= len(g.Map)) ||
		(entity.Pos.X+dx < 0 || entity.Pos.X+dx >= len(g.Map[entity.Pos.Y+dy])) {
		return
	}

	if g.QueryMap(newX, newY) != TileWalkable {
		return // Cannot move to unwalkable tile
	}

	entity.Pos.X += dx
	entity.Pos.Y += dy

	log.Printf("WORLD | %s moved to x: %v y: %v", entity.ID, entity.Pos.X, entity.Pos.Y)
}
