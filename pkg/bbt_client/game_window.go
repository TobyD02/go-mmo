package bbt_client

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/lipgloss"

	"github.com/tobyd02/go-mmo/pkg/client"
	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
)

type GameModel struct {
	client *client.GClient

	moveInputDir     [4]int // up, down, left, right
	interactInputDir [4]int // up, down, left, right

	chatInput  textinput.Model
	chatActive bool

	chatOutput string
}

type ConnectionErrorMsg struct {
	Err error
}

type GameTickMsg struct{}

func InitialModel(
	gClient *client.GClient,
) GameModel {

	chatInput := textinput.New()
	// chatInput.Placeholder = "Type a command"
	chatInput.SetValue("/")
	chatInput.CharLimit = 200

	return GameModel{
		client: gClient,

		chatInput:  chatInput,
		chatOutput: "",
	}
}

func tick() tea.Cmd {
	return tea.Tick(
		time.Millisecond,
		func(time.Time) tea.Msg {
			return GameTickMsg{}
		},
	)
}

func (m GameModel) Init() tea.Cmd {
	return tick()
}

func (m GameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case GameTickMsg:
		err := m.client.Update()
		if err != nil {
			return m, func() tea.Msg {
				return ConnectionErrorMsg{Err: err}
			}
		}

		moveX := m.moveInputDir[3] - m.moveInputDir[2]
		moveY := m.moveInputDir[1] - m.moveInputDir[0]

		if moveX != 0 || moveY != 0 {
			_ = m.client.Move(moveX, moveY)
		}

		m.moveInputDir = [4]int{}

		interactX := m.interactInputDir[3] - m.interactInputDir[2]
		interactY := m.interactInputDir[1] - m.interactInputDir[0]

		if interactX != 0 || interactY != 0 {
			m.client.InteractDirection(interactX, interactY)
		}

		m.interactInputDir = [4]int{}

		return m, tick()

	case tea.KeyPressMsg:
		if m.chatActive {
			switch msg.String() {
			case "enter":
				message := strings.TrimSpace(m.chatInput.Value())
				if message != "" {
					m.chatOutput = m.doCmd(message)
				}

				m.chatInput.SetValue("/")
				m.chatInput.Blur()
				m.chatActive = false

				return m, nil

			case "esc":
				m.chatInput.SetValue("/")
				m.chatInput.Blur()
				m.chatActive = false

				return m, nil
			}

			var cmd tea.Cmd
			m.chatInput, cmd = m.chatInput.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "/":
			m.chatActive = true
			return m, m.chatInput.Focus()

		case "w":
			m.moveInputDir[0] = 1

		case "s":
			m.moveInputDir[1] = 1

		case "a":
			m.moveInputDir[2] = 1

		case "d":
			m.moveInputDir[3] = 1

		case "up":
			m.interactInputDir[0] = 1

		case "down":
			m.interactInputDir[1] = 1

		case "left":
			m.interactInputDir[2] = 1

		case "right":
			m.interactInputDir[3] = 1

		}

	case ConnectionErrorMsg:
		return m, tea.Quit
	}

	return m, nil
}

func (m GameModel) View() tea.View {
	start := time.Now()
	var world strings.Builder
	var log strings.Builder
	var inventory strings.Builder

	viewportWidth := config.BBTClientViewportTilesX
	viewportHeight := config.BBTClientViewportTilesY

	clientPlayer := m.client.QuerySelf()
	if clientPlayer == nil {
		return tea.NewView("Loading...")
	}

	centerX := clientPlayer.Pos.X
	centerY := clientPlayer.Pos.Y

	startX := centerX - viewportWidth/2
	startY := centerY - viewportHeight/2

	endX := startX + viewportWidth
	endY := startY + viewportHeight

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {

			if !m.client.IsInBounds(x, y) {
				world.WriteString(" ")
				continue
			}

			players := m.client.QueryPlayers(x, y)
			interactableInstance := m.client.QueryInteractable(x, y)
			npcInstances := m.client.QueryNpcs(x, y)

			if len(players) > 0 {
				if players[m.client.ClientID] != nil {
					drawSelf(&world)
				} else {
					drawOther(&world)
				}

				continue
			}

			if interactableInstance != nil {
				drawInteractable(&world, interactableInstance, clientPlayer.ID)
				continue
			}

			if len(npcInstances) > 0 {
				for _, npcInstance := range npcInstances {
					drawNpc(&world, npcInstance, clientPlayer.ID)
					break
				}
				continue
			}

			drawTile(&world, m.client.QueryTile(x, y))
		}

		world.WriteString("\n")
	}

	for i := range m.client.Logs {
		if i < invHeight-1 {
			logMessage := m.client.Logs[len(m.client.Logs)-i-1]
			msg := fmt.Sprintf("%s | %s\n", logMessage.Scope, logMessage.Message)
			log.WriteString(msg)
		}
	}

	log.WriteString(fmt.Sprintf("CMD | %s", m.chatOutput))

	// Draw pos at top of inventory
	fmt.Fprintf(&inventory, "(%d, %d)\n", centerX, centerY)

	fmt.Fprintf(&inventory, "frametime: %-12s\n", time.Since(start))

	itemIDs := make([]string, 0, len(clientPlayer.Inventory))

	for itemID := range clientPlayer.Inventory {
		itemIDs = append(itemIDs, itemID)
	}

	sort.Strings(itemIDs)
	for _, itemID := range itemIDs {
		name := game.GetItemNameFromRegistry(itemID)
		amount := clientPlayer.Inventory[itemID]
		fmt.Fprintf(&inventory, "%-12s | %d\n", name, amount)
	}

	chatContent := "> "
	if m.chatActive {
		chatContent = m.chatInput.View()
	} else {
		chatContent = "Press /"
	}

	chatContent = chatStyle.Render(chatContent)

	worldContent := worldStyle.Render(world.String())
	inventoryContent := inventoryStyle.Render(inventory.String())
	logContent := logStyle.Render(log.String())
	bottomContent := lipgloss.JoinHorizontal(lipgloss.Top, inventoryContent, logContent)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		worldContent,
		bottomContent,
		chatContent,
	)

	return tea.NewView(gameStyle.Render(content))

	// return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, worldStyle.Render(world.String())+"\n"+logStyle.Render(log.String())))
}

func (m GameModel) doCmd(msg string) string {
	isCmd := strings.HasPrefix(msg, "/")
	if !isCmd {
		return "not a valid command"
	}

	parts := strings.Fields(msg)
	command := parts[0]
	args := parts[1:]

	switch command {
	case "/move":
		if len(args) != 2 {
			return "usage: /move <x> <y>"
		}

		x, err := strconv.Atoi(args[0])
		if err != nil {
			return "x must be a number"
		}

		y, err := strconv.Atoi(args[1])
		if err != nil {
			return "y must be a number"
		}

		if x < -1 || x > 1 || y < -1 || y > 1 {
			return "x and y must be -1, 0, or 1"
		}

		_ = m.client.Move(x, y)

		return fmt.Sprintf("Sent move command: dx %d, dy %d", x, y)
	}

	return ""
}
