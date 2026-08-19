package client

import (
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type GClient struct {
	conn     *websocket.Conn
	ClientID string
}

func NewGClient() *GClient {
	return &GClient{}
}

func (c *GClient) connectToServer(serverURI string) error {
	conn, _, err := websocket.DefaultDialer.Dial(
		serverURI+"/ws",
		nil,
	)
	if err != nil {
		return err
	}

	c.conn = conn
	return nil
}

func (c *GClient) Start(serverURI string, clientID string) (*game.GameWorld, error) {
	err := c.connectToServer(serverURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %s", err)
	}

	// At this point, connection is established. On failure, we need to close the connection
	success := false
	defer func() {
		if !success {
			c.conn.Close()
		}
	}()

	err = c.sendClientConnectedMessage(clientID)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with server: %s", err)
	}

	c.ClientID = clientID // set client ID now that we have authenticated

	msg, err := c.readMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to read message from server %s", err)
	}

	if msg.Type != messages.TServerInitialWorldStateMessage {
		return nil, fmt.Errorf("message received wasn't initial world state")
	}

	parsedData, err := messages.ParseGServerInitialWorldStateData(msg.Data)
	if err != nil {
		log.Fatalf("Failed to parse initial world state message")
	}

	success = true

	return parsedData.InitialWorldState, nil
}

func (c *GClient) StopAndCloseConnection() {
	c.conn.Close()
}

func (c *GClient) sendClientConnectedMessage(clientID string) error {
	return c.sendMessage(messages.NewGClientConnectedMessage(clientID))
}

func (c *GClient) SendMoveMessage(dx, dy int) error {
	return c.sendMessage(messages.NewGClientMoveMessage(dx, dy))
}

func (c *GClient) SendInteractMessage(interactableID string) error {
	return c.sendMessage(messages.NewGClientInteractMessage(interactableID))
}

func (c *GClient) sendMessage(msg *messages.GMessage, err error) error {
	if err != nil {
		return fmt.Errorf("failed to create message %s", err)
	}
	err = c.conn.WriteJSON(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to client %s", err)
	}

	return nil
}

func (c *GClient) readMessage() (*messages.GMessage, error) {
	_, message, err := c.conn.ReadMessage()

	if err != nil {
		return nil, fmt.Errorf("failed to read message from server %s", err)
	}

	return messages.ParseMessage(message)
}

func (c *GClient) ReadGameWorldDiff() (*game.GameWorldDiff, error) {
	msg, err := c.readMessage()

	if err != nil {
		return nil, fmt.Errorf("failed to read message %s", err)
	}

	if msg.Type != messages.TServerWorldDiffMessage {
		return nil, fmt.Errorf("didn't receive game world diff")
	}

	parsed, err := messages.ParseGServerWorldDiffData(msg.Data)

	if err != nil {
		return nil, fmt.Errorf("failed to parse game world diff message %s", err)
	}

	return parsed.WorldDiff, nil
}
