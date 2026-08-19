package bbt_client

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	game "github.com/tobyd02/golang-mmo/pkg/game_common"
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
)

var selfStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#499953")).Bold(true)
var otherStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#9c7d36")).Bold(true)
var interactableStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#74bce3")).Bold(true)
var interactableCooldownStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#1a3847")).Bold(true)

var worldStyle = lipgloss.NewStyle().Width(128).Height(32)
var logStyle = lipgloss.NewStyle().Width(128).Height(12).Border(lipgloss.NormalBorder())

func drawTile(b *strings.Builder, tile game.GameWorldTile) {
	b.WriteString(TileStyles[tile].Render(TileChars[tile]))
}

func drawSelf(b *strings.Builder) {
	b.WriteString(selfStyle.Render(selfChar))
}

func drawOther(b *strings.Builder) {
	b.WriteString(otherStyle.Render(otherChar))
}

func drawInteractable(b *strings.Builder) {
	b.WriteString(interactableStyle.Render(interactableChar))
}

func drawInteractableCooldown(b *strings.Builder) {
	b.WriteString(interactableCooldownStyle.Render(interactableCooldownChar))
}
