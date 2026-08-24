package map_editor

import (
	"fmt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type MEUi struct {
	getTileDrawer         func() *METileDrawer
	getNpcDrawer          func() *MENpcDrawer
	getInteractableDrawer func() *MEInteractableDrawer
	getMapEditor          func() *MEMapEditor
}

func NewMEUi(
	getTileDrawer func() *METileDrawer,
	getNpcDrawer func() *MENpcDrawer,
	getInteractableDrawer func() *MEInteractableDrawer,
	getMapDrawer func() *MEMapEditor,
) *MEUi {

	return &MEUi{
		getTileDrawer:         getTileDrawer,
		getInteractableDrawer: getInteractableDrawer,
		getNpcDrawer:          getNpcDrawer,
		getMapEditor:          getMapDrawer,
	}
}

func (u *MEUi) Draw(screen *ebiten.Image, editorMode MEMapEditorMode) {

	ebitenutil.DebugPrintAt(screen, "Left Click: Draw", 1000, 10)
	ebitenutil.DebugPrintAt(screen, "Right Click: Erase [Not Tile Mode]", 1000, 25)
	ebitenutil.DebugPrintAt(screen, "1: Next Entity (Floor Tile [Tile Mode])", 1000, 40)
	ebitenutil.DebugPrintAt(screen, "2: Wall Tile [Tile Mode]", 1000, 55)
	ebitenutil.DebugPrintAt(screen, "3: Spawn Tile [Tile Mode]", 1000, 70)
	ebitenutil.DebugPrintAt(screen, "]: Increase Brush Size", 1000, 85)
	ebitenutil.DebugPrintAt(screen, "[: Decrease Brush Size", 1000, 100)
	ebitenutil.DebugPrintAt(screen, "<TAB>: Next Mode", 1000, 115)
	ebitenutil.DebugPrintAt(screen, "Ctrl + S: Save", 1000, 130)

	if u.getMapEditor().lastErrorMessage != "" && time.Since(u.getMapEditor().lastErrorTime).Seconds() < 5 {
		ebitenutil.DebugPrint(screen, u.getMapEditor().lastErrorMessage)
	}

	currentMode := fmt.Sprintf("Mode: %s", modeToString[u.getMapEditor().mode])
	ebitenutil.DebugPrintAt(screen, currentMode, 10, 10)

	switch editorMode {
	case MEMapEditorMode_Tiles:
		currentTile := fmt.Sprintf("Selected Tile: %s", getTileName(u.getTileDrawer().activeTile))
		brushSize := fmt.Sprintf("Brush Size: %d", u.getTileDrawer().brushSize)

		ebitenutil.DebugPrintAt(screen, currentTile, 10, 25)
		ebitenutil.DebugPrintAt(screen, brushSize, 10, 40)

	case MEMapEditorMode_Npcs:
		currentNpc := fmt.Sprintf("Selected Npc: %s", u.getNpcDrawer().getActiveNpcName())
		brushSize := fmt.Sprintf("Brush Size: %d", u.getNpcDrawer().brushSize)

		ebitenutil.DebugPrintAt(screen, currentNpc, 10, 25)
		ebitenutil.DebugPrintAt(screen, brushSize, 10, 40)

	case MEMapEditorMode_Interactables:
		currentInteractable := fmt.Sprintf("Selected Interactable: %s", u.getInteractableDrawer().getActiveInteractableName())
		brushSize := fmt.Sprintf("Brush Size: %d", u.getInteractableDrawer().brushSize)

		ebitenutil.DebugPrintAt(screen, currentInteractable, 10, 25)
		ebitenutil.DebugPrintAt(screen, brushSize, 10, 40)
	}

}
