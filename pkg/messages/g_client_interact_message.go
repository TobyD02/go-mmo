package messages

import (
	"encoding/json"
	"fmt"
)

type GClientInteractMessage struct {
	InteractableID string `json:"interactable_id"`
}

func (m GClientInteractMessage) MessageName() string {
	return "interact action"
}

func NewGClientInteractMessage(interactableID string) (*GMessage, error) {

	data, err := json.Marshal(GClientInteractMessage{
		InteractableID: interactableID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create new client interact message %s", err)
	}

	return &GMessage{
		Type: TClientInteractMessage,
		Data: data,
	}, nil
}

func ParseGClientInteractMessageData(raw json.RawMessage) (*GClientInteractMessage, error) {
	var data GClientInteractMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
