package messages

import (
	"encoding/json"
	"fmt"

	"github.com/tobyd02/golang-mmo/pkg/game"
)

type GServerInitialWorldStateMessage struct {
	InitialWorldState *game.GameWorld `json:"initial_world_state"`
}

func (m *GServerInitialWorldStateMessage) MessageName() string {
	return "client connected"
}

func NewGServerInitialWorldStateMessage(gameWorld *game.GameWorld) (*GMessage, error) {
	data, err := json.Marshal(GServerInitialWorldStateMessage{
		InitialWorldState: gameWorld,
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
