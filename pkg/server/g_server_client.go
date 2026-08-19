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
	Messages   chan []byte
	Conn       *websocket.Conn
	writeMutex sync.Mutex
	ID         string
}

func NewGServerClient(conn *websocket.Conn) *GServerClient {
	clientMessages := make(chan []byte, 100)

	return &GServerClient{
		Messages: clientMessages,
		Conn:     conn,
		ID:       "",
	}
}

func (c *GServerClient) EstablishConnection() error {
	_, message, err := c.Conn.ReadMessage()

	if err != nil {
		return fmt.Errorf("failed to read connection message: %s", err)
	}

	clientConnectionMessage, err := messages.ParseMessage(message)

	if err != nil || clientConnectionMessage.Type != messages.TClientConnectedMessage {
		return fmt.Errorf("invalid connection message: %s", err)
	}

	connectionData, err := messages.ParseGClientConnectedData(clientConnectionMessage.Data)
	if err != nil {
		return fmt.Errorf("invalid connection message: %s", err)
	}

	c.ID = connectionData.ID
	return nil
}

func (c *GServerClient) ReadMessages() error {

	c.Conn.SetReadDeadline(time.Now().Add(10 * time.Second))

	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("client disconnected:", err)
			return err
		}

		c.Messages <- message
	}
}

func (c *GServerClient) WriteMessage(messageType int, data []byte) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()

	return c.Conn.WriteMessage(messageType, data)
}

func (c *GServerClient) PingLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		err := c.Conn.WriteControl(
			websocket.PingMessage,
			nil,
			time.Now().Add(time.Second),
		)

		if err != nil {
			log.Println("client ping failed:", c.ID)
			return
		}
	}
}

func (c *GServerClient) DrainMessages() map[messages.GMessageType]*messages.GMessage {
	msgs := make(map[messages.GMessageType]*messages.GMessage)

drain:
	for {
		select {
		case a := <-c.Messages:
			message, err := messages.ParseMessage(a)

			if err != nil {
				log.Printf("failed to decode client action: %s", err)
				break
			}

			msgs[message.Type] = message
		default:
			break drain
		}

	}

	return msgs
}
