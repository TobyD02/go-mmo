package messages

import (
	"encoding/json"
	"fmt"
)

type GClientAttackNpcMessage struct {
	NpcInstanceID string `json:"npc_instance_id"`
}

func (m GClientAttackNpcMessage) MessageName() string {
	return "interact action"
}

func NewGClientAttackNpcMessage(npcInstanceID string) (*GMessage, error) {

	data, err := json.Marshal(GClientAttackNpcMessage{
		NpcInstanceID: npcInstanceID,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create new client interact message %s", err)
	}

	return &GMessage{
		Type: TClientAttackNpcMessage,
		Data: data,
	}, nil
}

func ParseGClientAttackNpcMessageData(raw json.RawMessage) (*GClientAttackNpcMessage, error) {
	var data GClientAttackNpcMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
