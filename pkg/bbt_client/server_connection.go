// Package bbt_client - contains client specific logic
package bbt_client

import (
	"encoding/json"

	tea "charm.land/bubbletea/v2"
	"github.com/gorilla/websocket"
	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/messages"
)

type ServerConnection struct {
	Conn *websocket.Conn
}

func NewServerConnection(conn *websocket.Conn) *ServerConnection {
	return &ServerConnection{
		Conn: conn,
	}
}

func (s *ServerConnection) SendMoveAction(dx, dy int) tea.Cmd {
	return func() tea.Msg {
		return s.sendMessage(messages.NewGClientMoveMessage(dx, dy))
	}
}

func (s *ServerConnection) SendInteractAction(interactableID string) tea.Cmd {
	return func() tea.Msg {
		return s.sendMessage(messages.NewGClientInteractMessage(interactableID))
	}
}

func (s *ServerConnection) sendMessage(data *messages.GMessage, err error) tea.Msg {
	if err != nil {
		return ConnectionErrorMsg{
			Err: err,
		}
	}

	err = s.Conn.WriteJSON(data)

	if err != nil {
		return ConnectionErrorMsg{
			Err: err,
		}
	}

	return nil
}

func (s *ServerConnection) ReadGameWorldDiff() tea.Cmd {
	return func() tea.Msg {
		_, message, err := s.Conn.ReadMessage()

		if err != nil {
			return ConnectionErrorMsg{
				Err: err,
			}
		}

		var gameWorldDiff game.GameWorldDiff

		if err := json.Unmarshal(message, &gameWorldDiff); err != nil {
			return ConnectionErrorMsg{
				Err: err,
			}
		}

		return gameWorldDiff
	}
}
