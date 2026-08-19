// Package messages - contains the structs for all communications between client and server
package messages

import (
	"encoding/json"
)

type GMessageType int

const (
	TClientConnectedMessage GMessageType = iota
	TClientMoveMessage
	TClientInteractMessage
	TServerInitialWorldStateMessage
	TServerWorldDiffMessage
)

type GMessageData interface {
	MessageName() string
}

type GMessage struct {
	Type GMessageType    `json:"type"`
	Data json.RawMessage `json:"data"`
}

func ParseMessage(data []byte) (*GMessage, error) {
	var message GMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return nil, err
	}

	return &message, nil
}
