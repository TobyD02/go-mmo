package client

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type GClient struct {
	conn     *websocket.Conn
	ClientID string

	InboundMessages  chan []byte
	OutboundMessages chan []byte
}

func NewGClient() *GClient {
	return &GClient{
		InboundMessages:  make(chan []byte, 100),
		OutboundMessages: make(chan []byte, 100),
	}
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

	go c.ReadLoop()
	go c.WriteLoop()

	return parsedData.InitialWorldState, nil
}

func (c *GClient) WriteLoop() {
	for message := range c.OutboundMessages {
		err := c.conn.WriteMessage(
			messages.GWebsocketMessageType,
			message,
		)

		if err != nil {
			log.Printf("Failed to write message: %s", err)
			return
		}
	}
}

func (c *GClient) ReadLoop() {
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("failed to read message: %s", err)
			return
		}
		c.InboundMessages <- message
	}
}

func (c *GClient) StopAndCloseConnection() {
	c.conn.Close()
}

// sendClientConnectedMessage -- Doesn't use c.sendMessage since it must remain synchronous
func (c *GClient) sendClientConnectedMessage(clientID string) error {
	return c.sendMessageSync(messages.NewGClientConnectedMessage(clientID))
}

func (c *GClient) SendMoveMessage(dx, dy int) error {
	return c.sendMessage(messages.NewGClientMoveMessage(dx, dy))
}

func (c *GClient) SendInteractMessage(interactableID string) error {
	return c.sendMessage(messages.NewGClientInteractMessage(interactableID))
}

// sendMessageSync - Sends a message synchronously (doesn't use the write loop goroutine)
func (c *GClient) sendMessageSync(msg *messages.GMessage, err error) error {
	if err != nil {
		return fmt.Errorf("failed to create message %s", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %s", err)
	}

	err = c.conn.WriteMessage(messages.GWebsocketMessageType, data)
	if err != nil {
		return fmt.Errorf("failed to send message: %s", err)
	}

	return nil

}

func (c *GClient) sendMessage(msg *messages.GMessage, err error) error {
	if err != nil {
		return fmt.Errorf("failed to create message %s", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %s", err)
	}

	c.OutboundMessages <- data

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
	message := <-c.InboundMessages

	msg, err := messages.ParseMessage(message)
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
