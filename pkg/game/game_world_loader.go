package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tobyd02/go-mmo/pkg/util"
)

type GameWorldFile struct {
	Tiles         [][]int                `json:"tiles"`
	Npcs          map[string][]util.Vec2 `json:"npcs"`
	Interactables map[string][]util.Vec2 `json:"interactables"`
}

func LoadGameWorldFile(worldFilePath string) (GameWorldFile, error) {
	data, err := os.ReadFile(worldFilePath)
	if err != nil {
		return GameWorldFile{}, fmt.Errorf("failed to read world file path: %s", err)
	}

	var gameWorldFile GameWorldFile
	err = json.Unmarshal(data, &gameWorldFile)

	if err != nil {
		return GameWorldFile{}, fmt.Errorf("failed to unmarshal world file data: %s", err)
	}
	return gameWorldFile, nil
}

func LoadGameWorld(worldFilePath string) (*GameWorld, error) {
	gameWorldFile, err := LoadGameWorldFile(worldFilePath)
	if err != nil {
		return nil, err
	}

	height := len(gameWorldFile.Tiles)
	width := len(gameWorldFile.Tiles[0])
	gamePlayers := make(map[string]*GPlayer)

	spawn := util.Vec2{X: 0, Y: 0}

	return &GameWorld{
		Map:           loadTileLayer(gameWorldFile, &spawn),
		Players:       gamePlayers,
		Npcs:          loadNpcs(gameWorldFile),
		Interactables: loadInteractables(gameWorldFile),
		Width:         width, Height: height,
		SpawnPoint: spawn,
	}, nil
}

// Current limitation - only 1 entity can be assigned to each tile on world generation.
// Not that big of an issue tbh
func loadTileLayer(
	wf GameWorldFile,
	spawn *util.Vec2,
) [][]GameWorldTile {
	gameMap := make([][]GameWorldTile, 0, len(wf.Tiles))

	for y, row := range wf.Tiles {
		gameMapRow := make([]GameWorldTile, len(row))

		for x, tileID := range row {

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

func loadNpcs(wf GameWorldFile) map[string]*GNpcInstance {
	npcs := make(map[string]*GNpcInstance)

	for registryID, positions := range wf.Npcs {
		for _, pos := range positions {
			idInput := fmt.Sprintf("%s:%d:%d", registryID, pos.X, pos.Y)
			id := deterministicID(idInput)
			npcs[id] = NewGNpcInstance(id, registryID, pos.X, pos.Y)
		}
	}

	return npcs
}

func loadInteractables(wf GameWorldFile) map[string]*GInteractableInstance {
	interactables := make(map[string]*GInteractableInstance)

	occupiedPositions := make(map[util.Vec2]struct{})

	for registryID, positions := range wf.Interactables {
		for _, pos := range positions {

			// Only one interactable allowed per tile
			if _, exists := occupiedPositions[pos]; exists {
				panic(fmt.Sprintf(
					"multiple interactables at position %d,%d",
					pos.X,
					pos.Y,
				))
			}

			occupiedPositions[pos] = struct{}{}

			idInput := fmt.Sprintf(
				"%s:%d:%d",
				registryID,
				pos.X,
				pos.Y,
			)

			id := deterministicID(idInput)

			interactables[id] = NewGInteractableInstance(
				id,
				registryID,
				pos.X,
				pos.Y,
			)
		}
	}

	return interactables
}

func deterministicID(s string) string {
	hash := sha256.Sum256([]byte(s))

	return hex.EncodeToString(hash[:])
}
