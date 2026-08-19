package messages

import (
	"encoding/json"
	"fmt"

	"github.com/tobyd02/golang-mmo/pkg/game"
)

type GServerWorldDiffMessage struct {
	WorldDiff *game.GameWorldDiff `json:"world_diff"`
}

func (m *GServerWorldDiffMessage) MessageName() string {
	return "client connected"
}

func NewGServerWorldDiffMessage(worldDiff *game.GameWorldDiff) (*GMessage, error) {
	data, err := json.Marshal(GServerWorldDiffMessage{
		WorldDiff: worldDiff,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create server initial world state: %s", err)
	}

	return &GMessage{
		Type: TServerWorldDiffMessage,
		Data: data,
	}, nil
}

func ParseGServerWorldDiffData(raw json.RawMessage) (*GServerWorldDiffMessage, error) {
	var data GServerWorldDiffMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
