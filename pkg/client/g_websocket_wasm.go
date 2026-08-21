//go:build js

//^^ Tells compiler to only build if target platform is javascript
// We can have a multiple declarations of NewClientWebsocket() method - but since only one will be built depending on build target, we dont need to worry about it

package client

import (
	"context"

	"github.com/coder/websocket"
	"github.com/tobyd02/go-mmo/pkg/messages"
)

type WasmWebSocketClient struct {
	conn   *websocket.Conn
	ctx    context.Context
	cancel context.CancelFunc
}

func NewClientWebsocket() ClientWebsocket {
	ctx, cancel := context.WithCancel(context.Background())
	return &WasmWebSocketClient{
		ctx:    ctx,
		cancel: cancel,
	}
}

func (c *WasmWebSocketClient) Connect(uri string) error {
	conn, _, err := websocket.Dial(c.ctx, uri, nil)
	if err != nil {
		return err
	}

	conn.SetReadLimit(1 * 1024 * 1024) // 1MB

	c.conn = conn
	return nil
}

func (c *WasmWebSocketClient) ReadMessage() ([]byte, error) {
	_, message, err := c.conn.Read(c.ctx)
	return message, err
}

func (c *WasmWebSocketClient) WriteMessage(message []byte) error {
	msgType := websocket.MessageBinary
	if messages.GWebsocketMessageType == 1 { // TextMessage
		msgType = websocket.MessageText
	}
	return c.conn.Write(c.ctx, msgType, message)
}

func (c *WasmWebSocketClient) Close() error {
	c.cancel() // unblocks any in-flight Read
	return c.conn.Close(websocket.StatusNormalClosure, "")
}
