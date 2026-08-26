package messages

import (
	"encoding/json"
	"fmt"
	"maps"
	"reflect"

	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/util"
)

type GServerInitialWorldStateMessage struct {
	InitialWorldStateDiff *game.GameWorldDiff `json:"initial_world_state_diff"`
}

func (m *GServerInitialWorldStateMessage) MessageName() string {
	return "client connected"
}

// @todo - Improving initial world state
// At current, the save isn't massive, but thats because worlds dont currently generate with entities and interactables.
// Game World generation should be able to spawn interactables and npcs. Then when the diff is calculated, the resultant difference will only be
// for changed chunks.
// as server lifetime increases, this will likely be much greater. So in the future this will still be a bottleneck potentially
func NewGServerInitialWorldStateMessage(currentGameWorld *game.GameWorld, initialGameWorld *game.GameWorld) (*GMessage, error) {
	diff := game.GameWorldDiff{
		PlayersDiff:       make(map[string]*game.GPlayer),
		NpcsDiff:          make(map[string]*game.GNpcInstance),
		InteractablesDiff: make(map[string]*game.GInteractableInstance),
	}

	// players diff is always - since initial map includes none going to be all
	maps.Copy(diff.PlayersDiff, currentGameWorld.Players)

	// NPC diff
	for id, currentNPC := range currentGameWorld.Npcs {
		initialNPC, exists := initialGameWorld.Npcs[id]

		// NPC didn't exist initially, or has changed.
		if !exists || !reflect.DeepEqual(currentNPC, initialNPC) {
			diff.NpcsDiff[id] = currentNPC
		}
	}

	// NPCs that did exist but no longer
	for id := range initialGameWorld.Npcs {
		if _, exists := currentGameWorld.Npcs[id]; !exists {
			diff.NpcsDiff[id] = nil
		}
	}

	// Interactable diff
	for id, currentInteractable := range currentGameWorld.Interactables {
		initialInteractable, exists := initialGameWorld.Interactables[id]

		// Interactable didn't exist initially, or has changed.
		if !exists || !reflect.DeepEqual(currentInteractable, initialInteractable) {
			diff.InteractablesDiff[id] = currentInteractable
		}
	}

	// Interactables that did exist but no longer
	for id := range initialGameWorld.Interactables {
		if _, exists := currentGameWorld.Interactables[id]; !exists {
			diff.InteractablesDiff[id] = nil
		}
	}

	// Changed tiles - need to iterate through map, and if any tiles are different, amend them
	mapDiff := make(map[util.Vec2]game.GameWorldTile)
	for y, row := range currentGameWorld.Map {
		for x, tile := range row {
			if currentGameWorld.Map[y][x] != initialGameWorld.Map[y][x] {
				mapDiff[util.Vec2{X: x, Y: y}] = tile
			}
		}
	}

	data, err := json.Marshal(GServerInitialWorldStateMessage{
		InitialWorldStateDiff: &diff,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create server initial world state: %s", err)
	}

	return &GMessage{
		Type: TServerInitialWorldStateMessage,
		Data: data,
	}, nil
}

func ParseGServerInitialWorldStateData(raw json.RawMessage) (*GServerInitialWorldStateMessage, error) {
	var data GServerInitialWorldStateMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
