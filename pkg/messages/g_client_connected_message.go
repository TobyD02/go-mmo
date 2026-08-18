package messages

import (
	"encoding/json"
	"fmt"
)

type GClientConnectedMessage struct {
	ID string `json:"id"`
}

func (m *GClientConnectedMessage) MessageName() string {
	return "client connected"
}

func NewGClientConnectedMessage(id string) (*GMessage, error) {
	data, err := json.Marshal(GClientConnectedMessage{
		ID: id,
	})

	if err != nil {
		return nil, fmt.Errorf("Failed to create client connected message: %s", err)
	}

	return &GMessage{
		Type: TClientConnectedMessage,
		Data: data,
	}, nil
}

func ParseGClientConnectedData(raw json.RawMessage) (*GClientConnectedMessage, error) {
	var data GClientConnectedMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
