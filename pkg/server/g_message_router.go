package server

import (
	"fmt"
	"log"

	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type GMessageRouter struct {
	globalMessages [][]byte
	clientMessages map[string][][]byte
}

func NewGMessageRouter() *GMessageRouter {
	return &GMessageRouter{
		globalMessages: make([][]byte, 0),
		clientMessages: make(map[string][][]byte),
	}
}

func (m *GMessageRouter) pushGlobalMessage(msg []byte) {
	m.globalMessages = append(m.globalMessages, msg)
}

func (m *GMessageRouter) pushClientMessage(clientID string, msg []byte) {
	m.clientMessages[clientID] = append(m.clientMessages[clientID], msg)
}

func (m *GMessageRouter) Flush(
	currentClients map[string]*GServerClient,
	readOnlyClients map[string]*GServerClient,
) {
	for clientID, client := range currentClients {
		for _, globalMessage := range m.globalMessages {
			client.WriteMessage(globalMessage)
		}
		for _, clientMessage := range m.clientMessages[clientID] {
			client.WriteMessage(clientMessage)
		}
	}

	for _, client := range readOnlyClients {
		for _, globalMessage := range m.globalMessages {
			client.WriteMessage(globalMessage)
		}
	}

	m.globalMessages = m.globalMessages[:0]
	clear(m.clientMessages)
}

func (m *GMessageRouter) PushWorldDiffMessage(worldDiff *game.GameWorldDiff) error {
	msg, err := messages.NewGServerWorldDiffMessage(worldDiff)
	if err != nil {
		return fmt.Errorf("failed to generate world diff message")
	}

	encoded, err := msg.Encode()

	if err != nil {
		return fmt.Errorf("failed to generate world diff message: %s", err)
	}

	m.pushGlobalMessage(encoded)
	return nil
}

func (m *GMessageRouter) PushClientLogMessage(clientID string, scope string, message string) {
	msg, err := messages.NewGServerLogMessage(scope, message)

	if err != nil {
		log.Println("Failed to build client log message")
		return
	}

	data, err := msg.Encode()
	if err != nil {
		log.Println("Failed to encode client log message")
		return
	}

	m.pushClientMessage(clientID, data)
}

func (m *GMessageRouter) PushGlobalLogMessage(scope string, message string) {
	msg, err := messages.NewGServerLogMessage(scope, message)

	if err != nil {
		log.Println("Failed to build global log message")
		return
	}

	data, err := msg.Encode()
	if err != nil {
		log.Println("Failed to encode global log message")
		return
	}

	m.pushGlobalMessage(data)
}
