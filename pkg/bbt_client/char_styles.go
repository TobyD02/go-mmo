package bbt_client

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
)

var TileChars = map[game.GameWorldTile]string{
	game.TileFloor: "..",
	game.TileWall:  "██",
	game.TileSpawn: "SS",
}

var TileStyles = map[game.GameWorldTile]lipgloss.Style{
	game.TileFloor: lipgloss.NewStyle().Foreground(lipgloss.Color("#383838")).Bold(true),
	game.TileWall:  lipgloss.NewStyle().Foreground(lipgloss.Color("#919191")).Bold(true),
}

var (
	selfChar         = "MM"
	otherChar        = "@@"
	interactableChar = "II"
	npcChar          = "OO"
)

var selfStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#499953")).Bold(true)
var npcStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a8326b")).Bold(true)
var npcTargettingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fcae38")).Bold(true)
var otherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9c7d36")).Bold(true)
var interactableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#74bce3")).Bold(true)
var interactableOccupiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ec276")).Bold(true)
var interactableOccupiedOtherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c2a065")).Bold(true)
var interactableCooldownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a3847")).Bold(true)

// World viewport dimensions in terminal characters.
var worldWidth = config.ClientViewportTilesX * config.ClientTileWidth
var worldHeight = config.ClientViewportTilesY

// Inventory/log area below the world.
var invWidth = worldWidth / 4
var invHeight = config.ClientRows - worldHeight - 1 - 2

// -1 = world border
// -2 = chat height
var logWidth = worldWidth - invWidth
var logHeight = invHeight

// Chat spans the full viewport width.
var chatWidth = worldWidth
var chatHeight = 2

// Overall game dimensions.
var styleX = worldWidth
var styleY = config.ClientRows

var gameStyle = lipgloss.NewStyle().
	Width(styleX).
	Height(styleY).
	Border(lipgloss.DoubleBorder()).
	BorderTop(true).
	BorderBottom(true).
	BorderLeft(true).
	BorderRight(true)

var worldStyle = lipgloss.NewStyle().
	Width(worldWidth).
	Height(worldHeight).
	Border(lipgloss.NormalBorder(), false, false, true, false)

var inventoryStyle = lipgloss.NewStyle().
	Width(invWidth).
	Height(invHeight).
	Border(lipgloss.NormalBorder(), false, true, false, false)

var logStyle = lipgloss.NewStyle().
	Width(logWidth).
	Height(logHeight)

var chatStyle = lipgloss.NewStyle().
	Width(chatWidth).
	Height(chatHeight).
	Border(lipgloss.NormalBorder(), true, false, false, false)

func drawTile(b *strings.Builder, tile game.GameWorldTile) {
	b.WriteString(TileStyles[tile].Render(TileChars[tile]))
}

func drawSelf(b *strings.Builder) {
	b.WriteString(selfStyle.Render(selfChar))
}

func drawOther(b *strings.Builder) {
	b.WriteString(otherStyle.Render(otherChar))
}

func drawNpc(b *strings.Builder, npcInstance *game.GNpcInstance, clientID string) {

	char := getNpcChar(npcInstance.NpcID)

	if npcInstance.PlayerTargetID == clientID {
		b.WriteString(npcTargettingStyle.Render(char))
	} else {
		b.WriteString(npcStyle.Render(char))
	}
}

func drawInteractable(b *strings.Builder, interactableInstance *game.GInteractableInstance, clientID string) {
	char := getInteractableChar(interactableInstance.InteractableID)

	if interactableInstance.CurrentTickCooldown > 0 {
		b.WriteString(interactableCooldownStyle.Render(char))
		return
	}

	occupiedBy := interactableInstance.OccupiedBy
	switch occupiedBy {
	case "":
		b.WriteString(interactableStyle.Render(char))
	case clientID:
		b.WriteString(interactableOccupiedStyle.Render(char))
	default:
		b.WriteString(interactableOccupiedOtherStyle.Render(char))
	}
}

func getInteractableChar(interactableID string) string {
	switch interactableID {
	case "interactable.gold_deposit":
		return "GG"

	case "interactable.oak_tree":
		return "TT"

	default:
		return interactableChar
	}
}

func getNpcChar(npcID string) string {
	switch npcID {
	case "npc.chicken":
		return "cc"

	case "npc.cow":
		return "CC"

	default:
		return npcChar
	}
}
