package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"

	"github.com/tobyd02/go-mmo/pkg/util"
)

type GameWorldTileMapType int

const (
	TileMapTypeTile GameWorldTileMapType = iota
	TileMapTypeNpc
	TileMapTypeInteractable
)

type GameWorldTileMapTileData struct {
	Type                 GameWorldTileMapType
	UniqueTypeIdentifier any // i.e. type in registry, or integer identifier for tiles
}

func GenerateGameWorldTileMapData() map[int]GameWorldTileMapTileData {
	npcRegistry, err := GetNpcRegistry()
	if err != nil {
		panic(err)
	}

	interactableRegistry, err := GetInteractableRegistry()
	if err != nil {
		panic(err)
	}

	// Build tile map - start with internal tiles
	tileMapData := make(map[int]GameWorldTileMapTileData)
	for i := range GameWorldTileCount {
		tileMapData[i] = GameWorldTileMapTileData{
			Type:                 TileMapTypeTile,
			UniqueTypeIdentifier: i,
		}
	}

	// next - add entries for npcs
	for npcId := range npcRegistry {
		tileMapData[len(tileMapData)] = GameWorldTileMapTileData{
			Type:                 TileMapTypeNpc,
			UniqueTypeIdentifier: npcId,
		}
	}

	// finally - add interactables
	for interactableId := range interactableRegistry {
		tileMapData[len(tileMapData)] = GameWorldTileMapTileData{
			Type:                 TileMapTypeInteractable,
			UniqueTypeIdentifier: interactableId,
		}
	}

	return tileMapData
}

type GameWorldFile struct {
	Tiles    [][]int `json:"tiles"`
	Entities [][]int `json:"entities"`
}

func LoadGameWorld(worldFilePath string) (*GameWorld, error) {
	tileMapTileData := GenerateGameWorldTileMapData()

	data, err := os.ReadFile(worldFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read world file path: %s", err)
	}

	var gameWorldFile GameWorldFile
	err = json.Unmarshal(data, &gameWorldFile)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal world file data: %s", err)
	}

	height := len(gameWorldFile.Tiles)
	width := len(gameWorldFile.Tiles[0])

	gameNpcs := make(map[string]*GNpcInstance)
	gameInteractables := make(map[string]*GInteractableInstance)
	gamePlayers := make(map[string]*GPlayer)

	spawn := util.Vec2{X: 0, Y: 0}

	gameWorldMap := loadTileLayer(gameWorldFile, tileMapTileData, &spawn)
	loadEntityLayer(gameWorldFile, tileMapTileData, &gameNpcs, &gameInteractables)

	return &GameWorld{
		Map:           gameWorldMap,
		Players:       gamePlayers,
		Npcs:          gameNpcs,
		Interactables: gameInteractables,
		Width:         width, Height: height,
		SpawnPoint: spawn,
	}, nil
}

// Current limitation - only 1 entity can be assigned to each tile on world generation.
// Not that big of an issue tbh
func loadTileLayer(
	wf GameWorldFile,
	typeLookup map[int]GameWorldTileMapTileData,
	spawn *util.Vec2,
) [][]GameWorldTile {
	gameMap := make([][]GameWorldTile, 0, len(wf.Tiles))

	for y, row := range wf.Tiles {
		gameMapRow := make([]GameWorldTile, len(row))

		for x, tileMapID := range row {

			tileData, exists := typeLookup[tileMapID]
			if !exists {
				panic("Tile map contains non-existent tile type")
			}

			if tileData.Type != TileMapTypeTile {
				panic("tile map data contains non tile-map type")
			}

			// Ensure that the unique identifier is an int
			tileID, ok := tileData.UniqueTypeIdentifier.(int)
			if !ok {
				panic("tile map data contains non-int type identifier")
			}

			tile := GameWorldTile(tileID)
			if tile == TileSpawn {
				spawn.X = x
				spawn.Y = y
			}

			gameMapRow[x] = tile
		}

		gameMap = append(gameMap, gameMapRow)
	}

	return gameMap
}

func loadEntityLayer(
	wf GameWorldFile,
	typeLookup map[int]GameWorldTileMapTileData,
	npcs *map[string]*GNpcInstance,
	interactables *map[string]*GInteractableInstance,
) {
	gNpcs := *npcs
	gInteractables := *interactables

	rand := rand.New(rand.NewSource(0))

	for y, row := range wf.Entities {
		for x, tileMapID := range row {

			if wf.Tiles[y][x] == 1 {
				continue
			}

			r := rand.Intn(100)
			if r == 1 {
				idInput := fmt.Sprintf("%d:%s:%d:%d", TileMapTypeNpc, "npc.chicken", x, y)
				id := deterministicID(idInput)
				gNpcs[id] = NewGNpcInstance(id, "npc.chicken", x, y)
			} else if r == 2 {
				idInput := fmt.Sprintf("%d:%s:%d:%d", TileMapTypeInteractable, "interactable.oak_tree", x, y)
				id := deterministicID(idInput)
				gInteractables[id] = NewGInteractableInstance(id, "interactable.oak_tree", x, y)
			}

			continue

			if tileMapID == 0 { // for entities layer - 0 means empty/no entity
				continue

			}
			tileData, exists := typeLookup[tileMapID]
			if !exists {
				panic("Entities map layer contains non-existent tile type")
			}

			if tileData.Type != TileMapTypeInteractable && tileData.Type != TileMapTypeNpc {
				panic("Entities map layer contains a non-interactable non-npc type")
			}

			// Ensure that the unique identifier is an string
			registryID, ok := tileData.UniqueTypeIdentifier.(string)
			if !ok {
				panic("entities map layer contains non-string type identifier")
			}

			idInput := fmt.Sprintf("%d:%s:%d:%d", tileData.Type, registryID, x, y)
			id := deterministicID(idInput)

			if tileData.Type == TileMapTypeNpc {
				gNpcs[id] = NewGNpcInstance(id, registryID, x, y)
			} else if tileData.Type == TileMapTypeInteractable {
				gInteractables[id] = NewGInteractableInstance(id, registryID, x, y)
			}
		}
	}

}

func deterministicID(s string) string {
	hash := sha256.Sum256([]byte(s))

	return hex.EncodeToString(hash[:])
}
