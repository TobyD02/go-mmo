package bbt_client

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/tobyd02/go-mmo/pkg/client"
	"github.com/tobyd02/go-mmo/pkg/game"
)

type GameModel struct {
	gameWorld *game.GameWorld
	client    *client.GClient
}

type ConnectionErrorMsg struct {
	Err error
}

type GameTickMsg struct{}

func InitialModel(
	gameWorld *game.GameWorld,
	client *client.GClient,
) GameModel {
	return GameModel{
		gameWorld: gameWorld,
		client:    client,
	}
}

func tick() tea.Cmd {
	return tea.Tick(
		client.GClientTickSpeed,
		func(time.Time) tea.Msg {
			return GameTickMsg{}
		},
	)
}

func (m GameModel) Init() tea.Cmd {
	return tick()
}

func (m GameModel) interactDirection(dx, dy int) (tea.Model, tea.Cmd) {
	self := m.gameWorld.Players[m.client.ClientID]

	newX := self.Pos.X + dx
	newY := self.Pos.Y + dy

	npcInstance := m.gameWorld.QueryNpcInstanceAtPosition(newX, newY)
	interactableInstance := m.gameWorld.QueryInteractableInstanceAtPosition(newX, newY)

	if npcInstance != nil {
		return m, BBTSendAttackNpcMessage(m.client, npcInstance.ID)
	}

	if interactableInstance != nil {
		return m, BBTSendInteractMessage(m.client, interactableInstance.ID)
	}

	return m, nil
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case GameTickMsg:
		m.client.Update()

		diff, err := m.client.ReadGameWorldDiff()
		if err != nil {
			return m, func() tea.Msg {
				return ConnectionErrorMsg{Err: err}
			}
		}

		if diff != nil {
			updateWorld(m.gameWorld, diff)
		}

		_ = m.client.ProcessServerLogMessages()

		return m, tick()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "w":
			return m, BBTSendMoveMessage(m.client, 0, -1)

		case "s":
			return m, BBTSendMoveMessage(m.client, 0, 1)

		case "a":
			return m, BBTSendMoveMessage(m.client, -1, 0)

		case "d":
			return m, BBTSendMoveMessage(m.client, 1, 0)

		case "up":
			return m.interactDirection(0, -1)

		case "down":
			return m.interactDirection(0, 1)

		case "left":
			return m.interactDirection(-1, 0)

		case "right":
			return m.interactDirection(1, 0)
		}

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
	var inventory strings.Builder

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
			interactable := m.gameWorld.QueryInteractableInstanceAtPosition(x, y)
			npc := m.gameWorld.QueryNpcInstanceAtPosition(x, y)

			if len(players) > 0 {
				if players[m.client.ClientID] != nil {
					drawSelf(&world)
				} else {
					drawOther(&world)
				}

				continue
			}

			if interactable != nil {
				drawInteractable(&world, interactable, client.ID)
				continue
			}

			if npc != nil {
				drawNpc(&world, npc, client.ID)
				continue
			}

			drawTile(&world, m.gameWorld.Map[y][x])
		}

		world.WriteString("\n")
	}

	for i := range m.client.Logs {
		if i < 10 {
			logMessage := m.client.Logs[len(m.client.Logs)-i-1]
			msg := fmt.Sprintf("%s | %s\n", logMessage.Scope, logMessage.Message)
			log.WriteString(msg)
		}
	}

	// Draw pos at top of inventory
	fmt.Fprintf(&inventory, "x: %d | y : %d\n", centerX, centerY)

	itemIDs := make([]string, 0, len(client.Inventory))

	for itemID := range client.Inventory {
		itemIDs = append(itemIDs, itemID)
	}

	sort.Strings(itemIDs)
	for _, itemID := range itemIDs {
		name := game.GetItemNameFromRegistry(itemID)
		amount := client.Inventory[itemID]
		fmt.Fprintf(&inventory, "%s | %d\n", name, amount)
	}

	worldContent := worldStyle.Render(world.String())

	inventoryContent := inventoryStyle.Render(inventory.String())

	logContent := logStyle.Render(log.String())

	bottomContent := lipgloss.JoinHorizontal(lipgloss.Top, inventoryContent, logContent)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		worldContent,
		bottomContent,
	)

	return tea.NewView(gameStyle.Render(content))

	// return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, worldStyle.Render(world.String())+"\n"+logStyle.Render(log.String())))
}
