package messages

import (
	"encoding/json"
	"fmt"
)

type GServerLogMessage struct {
	Scope   string `json:"scope"`
	Message string `json:"message"`
}

func (m *GServerLogMessage) MessageName() string {
	return "client connected"
}

func NewGServerLogMessage(scope string, message string) (*GMessage, error) {
	data, err := json.Marshal(GServerLogMessage{
		Scope: scope, Message: message,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create server log message: %s", err)
	}

	return &GMessage{
		Type: TServerLogMessage,
		Data: data,
	}, nil
}

func ParseGServerLogData(raw json.RawMessage) (*GServerLogMessage, error) {
	var data GServerLogMessage
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
