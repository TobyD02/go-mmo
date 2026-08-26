package client

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/messages"
	"github.com/tobyd02/go-mmo/pkg/util"
)

type GClient struct {
	conn     ClientWebsocket
	ClientID string
	Logs     []*messages.GServerLogMessage
	logLimit int

	clientWorld *GClientWorld

	InboundMessages chan []byte

	OutboundMessages      map[messages.GMessageType][]byte
	outboundMessagesMutex sync.Mutex

	DrainedMessages map[messages.GMessageType][]*messages.GMessage

	tickCounter     int
	lastMessageTick map[messages.GMessageType]int

	readOnly bool
	done     chan struct{}
}

func NewGClient(readOnly bool) (*GClient, error) {
	logLimit := 100

	initialWorldState, err := game.NewGameWorld(config.GameWorldFilePath)
	if err != nil {
		return nil, fmt.Errorf("Failed to load initial game world state")
	}

	return &GClient{
		InboundMessages:  make(chan []byte, 100),
		OutboundMessages: make(map[messages.GMessageType][]byte, 100),
		Logs:             make([]*messages.GServerLogMessage, 0, logLimit),
		logLimit:         logLimit,

		clientWorld: NewGClientWorld(initialWorldState),

		DrainedMessages: make(map[messages.GMessageType][]*messages.GMessage),

		lastMessageTick: make(map[messages.GMessageType]int),
		tickCounter:     0,

		readOnly: readOnly,
		conn:     NewClientWebsocket(),
		done:     make(chan struct{}),
	}, nil
}

func (c *GClient) connectToServer(serverURI string) error {
	uri := serverURI + "/ws"
	if c.readOnly {
		uri = uri + "/ro"
	}

	return c.conn.Connect(uri)
}

func (c *GClient) Start(serverURI string, clientID string) error {
	err := c.connectToServer(serverURI)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %s", err)
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
		return fmt.Errorf("failed to authenticate with server: %s", err)
	}

	c.ClientID = clientID // set client ID now that we have authenticated

	msg, err := c.readMessage()
	if err != nil {
		return fmt.Errorf("failed to read message from server %s", err)
	}

	if msg.Type != messages.TServerInitialWorldStateMessage {
		return fmt.Errorf("message received wasn't initial world state %s", msg.Type)
	}

	parsedData, err := messages.ParseGServerInitialWorldStateData(msg.Data)
	if err != nil {
		return fmt.Errorf("Failed to parse initial world state message")
	}

	success = true

	go c.ReadLoop()

	if !c.readOnly {
		go c.WriteLoop()
	}

	c.clientWorld.ApplyWorldDiff(parsedData.InitialWorldStateDiff)

	return nil
}

func (c *GClient) WriteLoop() {
	if c.readOnly {
		return // cannot write in read only mode
	}

	ticker := time.NewTicker(config.ClientTickSpeed)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.flushOutboundMessages()
		case <-c.done:
			return
		}
	}
}

func (c *GClient) flushOutboundMessages() {
	c.outboundMessagesMutex.Lock()
	messagesToSend := c.OutboundMessages
	c.OutboundMessages = make(map[messages.GMessageType][]byte)
	c.outboundMessagesMutex.Unlock()

	for _, msg := range messagesToSend {
		if err := c.conn.WriteMessage(msg); err != nil {
			log.Printf("Failed to write message: %s", err)
			c.StopAndCloseConnection()
		}
	}

}

func (c *GClient) ReadLoop() {
	for {
		message, err := c.conn.ReadMessage()

		if err != nil {
			select {
			case <-c.done:
				// expected - we expected the close
			default:
				fmt.Printf("READ ERROR: %T: %v", err, err)
			}
			close(c.done)
			return
		}
		select {
		case c.InboundMessages <- message:
		case <-c.done:
			return

		}
	}
}

func (c *GClient) StopAndCloseConnection() {
	select {
	case <-c.done:
		// already closed
	default:
		close(c.done)
	}
	c.conn.Close()
}

// sendClientConnectedMessage -- Doesn't use c.sendMessage since it must remain synchronous
func (c *GClient) sendClientConnectedMessage(clientID string) error {
	return c.sendMessageSync(messages.NewGClientConnectedMessage(clientID))
}

func (c *GClient) moveMessage(dx, dy int) error {
	return c.queueMessage(messages.NewGClientMoveMessage(dx, dy))
}

func (c *GClient) interactMessage(interactableInstanceID string) error {
	return c.queueMessage(messages.NewGClientInteractMessage(interactableInstanceID))
}

func (c *GClient) attackMessage(npcInstanceID string) error {
	return c.queueMessage(messages.NewGClientAttackNpcMessage(npcInstanceID))
}

// sendMessageSync - Sends a message synchronously (doesn't use the write loop goroutine)
func (c *GClient) sendMessageSync(msg *messages.GMessage, err error) error {
	// Doesn't safe guard read only connections...

	if err != nil {
		return fmt.Errorf("failed to create message %s", err)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %s", err)
	}

	err = c.conn.WriteMessage(data)
	if err != nil {
		return fmt.Errorf("failed to send message: %s", err)
	}

	return nil

}

func (c *GClient) queueMessage(msg *messages.GMessage, err error) error {

	if c.readOnly {
		return fmt.Errorf("cannot send messages in read only mode")
	}

	if err != nil {
		return fmt.Errorf("failed to create message %s", err)
	}

	messageType := msg.Type

	lastTick, exists := c.lastMessageTick[messageType]

	if exists == true {
		if lastTick == c.tickCounter {
			return fmt.Errorf("already submitted a message of this type this tick")
		}
	}

	c.lastMessageTick[messageType] = c.tickCounter

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %s", err)
	}

	c.outboundMessagesMutex.Lock()
	c.OutboundMessages[messageType] = data
	c.outboundMessagesMutex.Unlock()

	// TODO - Add better interpolating. Its a bit rubbish atm
	//if messageType == messages.TClientMoveMessage {
	//	c.predictMovement(msg)
	//}

	return nil
}

func (c *GClient) readMessage() (*messages.GMessage, error) {
	message, err := c.conn.ReadMessage()

	if err != nil {
		return nil, fmt.Errorf("failed to read message from server %s", err)
	}

	return messages.ParseMessage(message)
}

func (c *GClient) ReadGameWorldDiff() (*game.GameWorldDiff, error) {
	msg := c.popDrainedMessage(messages.TServerWorldDiffMessage)
	// If no game world diff
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
		return nil
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
func (c *GClient) Update() error {
	c.drainMessages()

	for {
		diff, err := c.ReadGameWorldDiff()
		if err != nil {
			return err
		}

		if diff == nil {
			break
		}

		c.clientWorld.ApplyWorldDiff(diff)
		// If we received a diff, then the server has ticked
		c.tickCounter++
	}

	self := c.QuerySelf()
	if self != nil {
		c.clientWorld.BuildPathFinder(self.Pos)
	}

	err := c.ProcessServerLogMessages()
	if err != nil {
		return err
	}

	return nil
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

func (c *GClient) Move(dx, dy int) error {
	return c.moveMessage(dx, dy)
}

func (c *GClient) Interact(dx, dy int) {
	self := c.QuerySelf()

	newX := self.Pos.X + dx
	newY := self.Pos.Y + dy

	npcInstances := c.clientWorld.QueryNpcInstancesAtPosition(newX, newY)
	interactableInstance := c.clientWorld.QueryInteractableInstanceAtPosition(newX, newY)

	for _, npcInstance := range npcInstances {
		_ = c.attackMessage(npcInstance.ID) // attack the first found and then return
		return
	}

	if interactableInstance != nil {
		_ = c.interactMessage(interactableInstance.ID)
		return
	}

	return
}

func (c *GClient) predictMovement(msg *messages.GMessage) error {
	movement, err := messages.ParseGClientMoveMessageData(msg.Data)
	if err != nil {
		return err
	}

	c.clientWorld.PredictClientMovement(movement.Dx, movement.Dy, c.ClientID)
	return nil
}

// ---- Wrappers for client world
func (c *GClient) QuerySelf() *game.GPlayer {
	return c.clientWorld.QueryPlayer(c.ClientID)
}

func (c *GClient) QueryInteractable(x, y int) *game.GInteractableInstance {
	return c.clientWorld.QueryInteractableInstanceAtPosition(x, y)
}

func (c *GClient) QueryPlayers(x, y int) map[string]*game.GPlayer {
	return c.clientWorld.QueryPlayersAtPosition(x, y)
}

func (c *GClient) QueryNpcs(x, y int) map[string]*game.GNpcInstance {
	return c.clientWorld.QueryNpcInstancesAtPosition(x, y)
}

func (c *GClient) QueryTile(x, y int) game.GameWorldTile {
	return c.clientWorld.QueryMap(x, y)
}

func (c *GClient) IsInBounds(x, y int) bool {
	return c.clientWorld.IsInBounds(x, y)
}

func (c *GClient) GetWorldDimensions() (width, height int) {
	return c.clientWorld.gameWorld.Width, c.clientWorld.gameWorld.Height
}

func (c *GClient) GetPathTo(target util.Vec2) []util.Vec2 {
	self := c.QuerySelf()
	if self == nil {
		return nil
	}

	// Cannot get path if out of bounds
	if !c.IsInBounds(target.X, target.Y) {
		return nil
	}

	return c.clientWorld.GetPath(self.Pos, target)
}

func (c *GClient) GetMovesTo(target util.Vec2) []util.Vec2 {
	path := c.GetPathTo(target)
	if path == nil {
		return nil
	}

	return c.clientWorld.PathToMoves(path)
}

func (c *GClient) PathToMoves(path []util.Vec2) []util.Vec2 {
	return c.clientWorld.PathToMoves(path)
}
