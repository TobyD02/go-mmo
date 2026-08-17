// Package server - clients, channels, etc... for running the game server
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

type GClient struct {
	Actions chan []byte
	Conn    *websocket.Conn
	ID      string
}

func NewGClient(conn *websocket.Conn) *GClient {
	clientActions := make(chan []byte, 100)

	return &GClient{
		clientActions,
		conn,
		"",
	}
}

func (c *GClient) EstablishConnection() error {
	_, message, err := c.Conn.ReadMessage()

	if err != nil {
		return fmt.Errorf("failed to read connection message: %s", err)
	}

	var clientConnection GClientConnection

	err = json.Unmarshal(message, &clientConnection)

	if err != nil {
		return fmt.Errorf("Invalid connection message: ", err)
	}

	c.ID = clientConnection.ID

	return nil
}

func (c *GClient) ReadActions() error {

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

		c.Actions <- message
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

func (c *GClient) DrainActions() []GClientAction {
	var actions []GClientAction

drain:
	for {
		select {
		case a := <-c.Actions:
			var action GClientAction
			err := json.Unmarshal(a, &action)

			if err != nil {
				log.Printf("failed to decode client action: %s", err)
				break
			}

			actions = append(actions, action)
		default:
			break drain
		}

	}

	return actions

}
