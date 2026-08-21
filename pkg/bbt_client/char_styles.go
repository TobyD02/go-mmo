package bbt_client

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tobyd02/golang-mmo/pkg/game"
)

var TileChars = map[game.GameWorldTile]string{
	game.TileWalkable: "..",
	game.TileWall:     "██",
}

var TileStyles = map[game.GameWorldTile]lipgloss.Style{
	game.TileWalkable: lipgloss.NewStyle().Foreground(lipgloss.Color("#383838")).Bold(true),
	game.TileWall:     lipgloss.NewStyle().Foreground(lipgloss.Color("#919191")).Bold(true),
}

var (
	selfChar                 = "MM"
	otherChar                = "@@"
	interactableChar         = "II"
	interactableCooldownChar = "||"
	npcChar                  = "OO"
)

var selfStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#499953")).Bold(true)
var npcStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#a8326b")).Bold(true)
var npcTargettingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#fcae38")).Bold(true)
var otherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9c7d36")).Bold(true)
var interactableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#74bce3")).Bold(true)
var interactableOccupiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#8ec276")).Bold(true)
var interactableOccupiedOtherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#c2a065")).Bold(true)
var interactableCooldownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a3847")).Bold(true)

var gameStyle = lipgloss.NewStyle().Width(128).Height(40).Border(lipgloss.NormalBorder())
var worldStyle = lipgloss.NewStyle().Width(128).Height(28)
var logStyle = lipgloss.NewStyle().Width(96).Height(12).Border(lipgloss.NormalBorder(), true, false, false, false)
var inventoryStyle = lipgloss.NewStyle().Width(32).Height(12).Border(lipgloss.NormalBorder(), true, true, false, false)

func drawTile(b *strings.Builder, tile game.GameWorldTile) {
	b.WriteString(TileStyles[tile].Render(TileChars[tile]))
}

func drawSelf(b *strings.Builder) {
	b.WriteString(selfStyle.Render(selfChar))
}

func drawOther(b *strings.Builder) {
	b.WriteString(otherStyle.Render(otherChar))
}

func drawNpc(b *strings.Builder, npcPlayerTargetID string, clientID string) {

	if npcPlayerTargetID == clientID {
		b.WriteString(npcTargettingStyle.Render(npcChar))
	} else {
		b.WriteString(npcStyle.Render(npcChar))
	}
}

func drawInteractable(b *strings.Builder, occupiedBy string, clientID string) {
	if occupiedBy == "" {
		b.WriteString(interactableStyle.Render(interactableChar))
	} else if occupiedBy == clientID {
		b.WriteString(interactableOccupiedStyle.Render(interactableChar))
	} else {
		b.WriteString(interactableOccupiedOtherStyle.Render(interactableChar))
	}
}

func drawInteractableCooldown(b *strings.Builder) {
	b.WriteString(interactableCooldownStyle.Render(interactableCooldownChar))
}
