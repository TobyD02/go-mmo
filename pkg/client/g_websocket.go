package client

type ClientWebsocket interface {
	Connect(uri string) error
	ReadMessage() ([]byte, error)
	WriteMessage(message []byte) error
	Close() error
}
