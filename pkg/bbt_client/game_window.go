package bbt_client

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/tobyd02/golang-mmo/pkg/client"
	"github.com/tobyd02/golang-mmo/pkg/game"
)

type GameModel struct {
	gameWorld *game.GameWorld
	client    *client.GClient
}

type ConnectionErrorMsg struct {
	Err error
}

func InitialModel(
	gameWorld *game.GameWorld,
	client *client.GClient,
) GameModel {
	return GameModel{
		gameWorld: gameWorld,
		client:    client,
	}
}

func (m GameModel) Init() tea.Cmd {
	teaCmd := BBTReadGameWorldDiff(m.client)

	return teaCmd
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "w":
			return m, BBTSendMoveMessage(m.client, 0, -1)

		case "down", "s":
			return m, BBTSendMoveMessage(m.client, 0, 1)

		case "left", "a":
			return m, BBTSendMoveMessage(m.client, -1, 0)

		case "right", "d":
			return m, BBTSendMoveMessage(m.client, 1, 0)

		case "space":
			self := m.gameWorld.Players[m.client.ClientID]
			interactable := m.gameWorld.QueryInteractableAtPosition(self.Pos.X-1, self.Pos.Y)

			if interactable != nil {
				return m, BBTSendInteractMessage(m.client, interactable.ID)
			}

		}

	case *game.GameWorldDiff:
		updateWorld(m.gameWorld, msg)

		// IMPORTANT:
		// Keep listening for the next server update.
		return m, BBTReadGameWorldDiff(m.client)

	case ConnectionErrorMsg:
		return m, tea.Quit
	}

	return m, nil
}

func updateWorld(
	world *game.GameWorld,
	diff *game.GameWorldDiff,
) {
	world.ApplyDiff(diff)
}

func (m GameModel) View() tea.View {
	var world strings.Builder
	var log strings.Builder

	viewportWidth := 64
	viewportHeight := 32

	client := m.gameWorld.Players[m.client.ClientID]
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
				if players[m.client.ClientID] != nil {
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

	worldContent := worldStyle.Render(world.String())
	logContent := logStyle.Render(log.String())
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		worldContent,
		logContent,
	)

	return tea.NewView(gameStyle.Render(content))

	// return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, worldStyle.Render(world.String())+"\n"+logStyle.Render(log.String())))
}
