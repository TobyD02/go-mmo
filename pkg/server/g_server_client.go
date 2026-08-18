// Package server - clients, channels, etc... for running the game server
package server

import (
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type GClient struct {
	Messages chan []byte
	Conn     *websocket.Conn
	ID       string
}

func NewGClient(conn *websocket.Conn) *GClient {
	clientMessages := make(chan []byte, 100)

	return &GClient{
		clientMessages,
		conn,
		"",
	}
}

func (c *GClient) EstablishConnection() error {
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

func (c *GClient) ReadMessages() error {

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

func (c *GClient) PingLoop() {
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

func (c *GClient) DrainMessages() []*messages.GMessage {
	var msgs []*messages.GMessage

drain:
	for {
		select {
		case a := <-c.Messages:
			message, err := messages.ParseMessage(a)

			if err != nil {
				log.Printf("failed to decode client action: %s", err)
				break
			}

			msgs = append(msgs, message)
		default:
			break drain
		}

	}

	return msgs
}
