package client

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

const GClientTickSpeed = time.Millisecond * 10

type GClient struct {
	conn     *websocket.Conn
	ClientID string
	Logs     []*messages.GServerLogMessage
	logLimit int

	InboundMessages  chan []byte
	OutboundMessages chan []byte

	DrainedMessages map[messages.GMessageType][]*messages.GMessage
}

func NewGClient() *GClient {
	logLimit := 100

	return &GClient{
		InboundMessages:  make(chan []byte, 100),
		OutboundMessages: make(chan []byte, 100),
		Logs:             make([]*messages.GServerLogMessage, 0, logLimit),
		logLimit:         logLimit,

		DrainedMessages: make(map[messages.GMessageType][]*messages.GMessage),
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
			fmt.Printf("READ ERROR: %T: %v", err, err)
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
	msg := c.popDrainedMessage(messages.TServerWorldDiffMessage)
	if msg == nil {
		return nil, nil
	}

	parsed, err := messages.ParseGServerWorldDiffData(msg.Data)

	if err != nil {
		return nil, fmt.Errorf("failed to parse game world diff message %s", err)
	}

	return parsed.WorldDiff, nil
}

func (c *GClient) ProcessServerLogMessages() error {
	logs := c.popAllDrainedMessages(messages.TServerLogMessage)
	if logs == nil {
		return fmt.Errorf("no logs to retrieve")
	}

	for _, log := range logs {
		parsed, err := messages.ParseGServerLogData(log.Data)
		if err != nil {
			continue
		}

		c.Logs = append(c.Logs, parsed)
	}

	// If logs exceeds log limit, cut it down
	if len(c.Logs) > c.logLimit {
		c.Logs = c.Logs[len(c.Logs)-c.logLimit:]
	}

	return nil
}

// Update - Must be called at the start of every client tick
func (c *GClient) Update() {
	c.drainMessages()
}

func (c *GClient) drainMessages() {
drain:
	for {
		select {
		case rawMessage, ok := <-c.InboundMessages:

			if !ok {
				break drain
			}

			message, err := messages.ParseMessage(rawMessage)

			if err != nil {
				log.Printf(
					"failed to decode server message: %s",
					err,
				)

				continue
			}

			c.DrainedMessages[message.Type] = append(
				c.DrainedMessages[message.Type], message)

		default:
			break drain
		}
	}
}

func (c *GClient) popDrainedMessage(messageType messages.GMessageType) *messages.GMessage {
	msgs := c.DrainedMessages[messageType]

	if len(msgs) == 0 {
		return nil
	}

	msg := msgs[0]

	if len(msgs) == 1 {
		delete(c.DrainedMessages, messageType)
	} else {
		c.DrainedMessages[messageType] = msgs[1:]
	}

	return msg
}

func (c *GClient) popAllDrainedMessages(messageType messages.GMessageType) []*messages.GMessage {
	msgs := c.DrainedMessages[messageType]

	delete(c.DrainedMessages, messageType)

	return msgs
}
