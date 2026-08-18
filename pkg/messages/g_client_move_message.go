package messages

import (
	"encoding/json"
	"fmt"
)

type GClientMoveMessage struct {
	Dx int `json:"dx"`
	Dy int `json:"dy"`
}

func (m GClientMoveMessage) MessageName() string {
	return "move action"
}

func NewGClientMoveMessage(dx, dy int) (*GMessage, error) {

	data, err := json.Marshal(GClientMoveMessage{
		Dx: dx, Dy: dy,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create new client move action %s", err)
	}

	return &GMessage{
		Type: TClientMoveMessage,
		Data: data,
	}, nil
}

func ParseGClientMoveMessageData(raw json.RawMessage) (*GClientMoveMessage, error) {
	var data GClientMoveMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
