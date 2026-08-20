// Package server - clients, channels, etc... for running the game server
package server

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type GServerClient struct {
	InboundMessages  chan []byte
	OutboundMessages chan []byte

	Conn *websocket.Conn
	ID   string

	done      chan struct{}
	closeOnce sync.Once
}

func NewGServerClient(conn *websocket.Conn) *GServerClient {
	return &GServerClient{
		InboundMessages:  make(chan []byte, 100),
		OutboundMessages: make(chan []byte, 100),
		Conn:             conn,
		ID:               "",
		done:             make(chan struct{}),
	}
}

func (c *GServerClient) Close() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.Conn.Close()
	})
}

func (c *GServerClient) WriteLoop() {
	for {
		select {
		case message := <-c.OutboundMessages:
			err := c.Conn.WriteMessage(
				messages.GWebsocketMessageType,
				message,
			)

			if err != nil {
				log.Printf(
					"failed to write to client %s: %s",
					c.ID,
					err,
				)

				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *GServerClient) EstablishConnection() error {
	_, message, err := c.Conn.ReadMessage()

	if err != nil {
		return fmt.Errorf(
			"failed to read connection message: %s",
			err,
		)
	}

	clientConnectionMessage, err := messages.ParseMessage(message)

	if err != nil ||
		clientConnectionMessage.Type != messages.TClientConnectedMessage {
		return fmt.Errorf(
			"invalid connection message: %s",
			err,
		)
	}

	connectionData, err := messages.ParseGClientConnectedData(
		clientConnectionMessage.Data,
	)

	if err != nil {
		return fmt.Errorf(
			"invalid connection message: %s",
			err,
		)
	}

	c.ID = connectionData.ID

	return nil
}

func (c *GServerClient) ReadLoop() error {
	c.Conn.SetReadDeadline(
		time.Now().Add(10 * time.Second),
	)

	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(
			time.Now().Add(10 * time.Second),
		)

		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()

		if err != nil {
			log.Println(
				"client disconnected:",
				err,
			)

			return err
		}

		select {
		case c.InboundMessages <- message:

		case <-c.done:
			return nil
		}
	}
}

func (c *GServerClient) WriteMessage(data []byte) {
	select {
	case c.OutboundMessages <- data:
	case <-c.done:
	default:
		fmt.Println("overflowed outbound messages", len(c.OutboundMessages))
	}
}

func (c *GServerClient) PingLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := c.Conn.WriteControl(
				websocket.PingMessage,
				nil,
				time.Now().Add(time.Second),
			)

			if err != nil {
				log.Println(
					"client ping failed:",
					c.ID,
				)

				return
			}

		case <-c.done:
			return
		}
	}
}

func (c *GServerClient) DrainMessages() map[messages.GMessageType]*messages.GMessage {
	msgs := make(map[messages.GMessageType]*messages.GMessage)

drain:
	for {
		select {
		case rawMessage := <-c.InboundMessages:
			message, err := messages.ParseMessage(rawMessage)

			if err != nil {
				log.Printf(
					"failed to decode client action: %s",
					err,
				)

				continue
			}

			msgs[message.Type] = message

		default:
			break drain
		}
	}

	return msgs
}
