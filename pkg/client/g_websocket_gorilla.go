//go:build !js

//^^ Tells compiler to only build if target platform is not javascript
// We can have a multiple declarations of NewClientWebsocket() method - but since only one will be built depending on build target, we dont need to worry about it

package client

import (
	"github.com/gorilla/websocket"
	"github.com/tobyd02/go-mmo/pkg/messages"
)

type GorillaWebSocketClient struct {
	conn *websocket.Conn
}

func NewClientWebsocket() ClientWebsocket {
	return &GorillaWebSocketClient{}
}

func (c *GorillaWebSocketClient) Connect(uri string) error {
	conn, _, err := websocket.DefaultDialer.Dial(
		uri,
		nil,
	)
	if err != nil {
		return err
	}

	c.conn = conn

	return nil
}

func (c *GorillaWebSocketClient) ReadMessage() ([]byte, error) {
	_, message, err := c.conn.ReadMessage()
	return message, err
}

func (c *GorillaWebSocketClient) WriteMessage(
	message []byte,
) error {
	return c.conn.WriteMessage(
		messages.GWebsocketMessageType,
		message,
	)
}

func (c *GorillaWebSocketClient) Close() error {
	return c.conn.Close()
}
