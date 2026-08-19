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
	clientId   string
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
	clientId string,
) GameModel {
	return GameModel{
		gameWorld:  gameWorld,
		serverConn: NewServerConnection(conn),
		clientId:   clientId,
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

		case "space":
			self := m.gameWorld.Players[m.clientId]
			interactable := m.gameWorld.Clone().QueryInteractableAtPosition(self.Pos.X-1, self.Pos.Y)

			if interactable != nil {
				return m, m.serverConn.SendInteractAction(interactable.ID)
			}

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
	var world strings.Builder
	var log strings.Builder

	viewportWidth := 64
	viewportHeight := 32

	client := m.gameWorld.Players[m.clientId]
	if client == nil {
		return tea.NewView("Loading...")
	}

	centerX := client.Pos.X
	centerY := client.Pos.Y

	startX := centerX - viewportWidth/2
	startY := centerY - viewportHeight/2

	endX := startX + viewportWidth
	endY := startY + viewportHeight

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {

			// Outside the world
			if y < 0 || y >= m.gameWorld.Height ||
				x < 0 || x >= m.gameWorld.Width {
				world.WriteString("  ")
				continue
			}

			players := m.gameWorld.QueryPlayersAtPosition(x, y)
			interactable := m.gameWorld.QueryInteractableAtPosition(x, y)

			if len(players) > 0 {
				if players[m.clientId] != nil {
					drawSelf(&world)
				} else {
					drawOther(&world)
				}

				continue
			}

			if interactable != nil {
				if interactable.CurrentTickCooldown <= 0 {
					drawInteractable(&world)
				} else {
					drawInteractableCooldown(&world)
				}
				continue
			}

			drawTile(&world, m.gameWorld.Map[y][x])
		}

		world.WriteString("\n")
	}

	return tea.NewView(worldStyle.Render(world.String()) + "\n" + logStyle.Render(log.String()))
}
