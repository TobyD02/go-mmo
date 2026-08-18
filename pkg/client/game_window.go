package client

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/gorilla/websocket"

	game "github.com/tobyd02/golang-mmo/pkg/game_common"
)

type GameModel struct {
	gameWorld  *game.GameWorld
	serverConn *ServerConnection
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
		gameWorld:  gameWorld,
		serverConn: NewServerConnection(conn),
	}
}

func (m GameModel) Init() tea.Cmd {
	return m.serverConn.ReadGameWorldDiff()
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "w":
			return m, m.serverConn.SendMoveAction(0, -1)

		case "down", "s":
			return m, m.serverConn.SendMoveAction(0, 1)

		case "left", "a":
			return m, m.serverConn.SendMoveAction(-1, 0)

		case "right", "d":
			return m, m.serverConn.SendMoveAction(1, 0)
		}

	case game.GameWorldDiff:
		updateWorld(m.gameWorld, msg)

		// IMPORTANT:
		// Keep listening for the next server update.
		return m, m.serverConn.ReadGameWorldDiff()

	case ConnectionErrorMsg:
		return m, tea.Quit
	}

	return m, nil
}

func updateWorld(
	world *game.GameWorld,
	diff game.GameWorldDiff,
) {
	world.ApplyDiff(diff)
}

func (m GameModel) View() tea.View {
	var b strings.Builder

	for _, row := range m.gameWorld.Map {
		for _, tile := range row {
			switch tile {
			case game.TileBlank:
				b.WriteString("..")

			case game.TileEntity:
				b.WriteString("██")
			}
		}

		b.WriteString("\n")
	}

	return tea.NewView(b.String())
}
