package client

import (
	"encoding/json"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/gorilla/websocket"

	"github.com/tobyd02/golang-mmo/pkg/game"
	"github.com/tobyd02/golang-mmo/pkg/server"
)

type GameModel struct {
	gameWorld *game.GameWorld
	conn      *websocket.Conn
}

type WorldStateMsg map[string]struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type ConnectionErrorMsg struct {
	Err error
}

func InitialModel(
	gameWorld *game.GameWorld,
	conn *websocket.Conn,
) GameModel {
	return GameModel{
		gameWorld: gameWorld,
		conn:      conn,
	}
}

func (m GameModel) Init() tea.Cmd {
	return readWorldState(m.conn)
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "w":
			return m, sendAction(m.conn, 0, -1)

		case "down", "s":
			return m, sendAction(m.conn, 0, 1)

		case "left", "a":
			return m, sendAction(m.conn, -1, 0)

		case "right", "d":
			return m, sendAction(m.conn, 1, 0)
		}

	case WorldStateMsg:
		updateWorld(m.gameWorld, msg)

		// IMPORTANT:
		// Keep listening for the next server update.
		return m, readWorldState(m.conn)

	case ConnectionErrorMsg:
		return m, tea.Quit
	}

	return m, nil
}

func updateWorld(
	world *game.GameWorld,
	state WorldStateMsg,
) {
	// Clear the existing map.
	for y := range world.Map {
		for x := range world.Map[y] {
			world.Map[y][x] = game.TileBlank
		}
	}

	// Draw whatever the server says exists.
	for _, position := range state {
		if position.Y < 0 ||
			position.Y >= len(world.Map) {
			continue
		}

		if position.X < 0 ||
			position.X >= len(world.Map[position.Y]) {
			continue
		}

		world.Map[position.Y][position.X] = game.TilePlayer
	}
}

func readWorldState(conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		_, message, err := conn.ReadMessage()

		if err != nil {
			return ConnectionErrorMsg{
				Err: err,
			}
		}

		var worldState WorldStateMsg

		if err := json.Unmarshal(message, &worldState); err != nil {
			return ConnectionErrorMsg{
				Err: err,
			}
		}

		return worldState
	}
}

func sendAction(conn *websocket.Conn, dx, dy int) tea.Cmd {
	return func() tea.Msg {
		err := conn.WriteJSON(server.GClientAction{
			Dx: dx,
			Dy: dy,
		})

		if err != nil {
			return ConnectionErrorMsg{
				Err: err,
			}
		}

		return nil
	}
}

func (m GameModel) View() tea.View {
	var b strings.Builder

	for _, row := range m.gameWorld.Map {
		for _, tile := range row {
			switch tile {
			case game.TileBlank:
				b.WriteString("·")

			case game.TilePlayer:
				b.WriteString("@")
			}
		}

		b.WriteString("\n")
	}

	return tea.NewView(b.String())
}
