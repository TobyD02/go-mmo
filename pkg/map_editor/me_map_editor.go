package map_editor

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
)

const METileSize = 10

type MEMapEditorMode int

const mapEditorModeCount int = 3

const (
	MEMapEditorMode_Tiles MEMapEditorMode = iota
	MEMapEditorMode_Npcs
	MEMapEditorMode_Interactables
)

var modeToString = map[MEMapEditorMode]string{
	MEMapEditorMode_Tiles:         "Tiles",
	MEMapEditorMode_Npcs:          "Npcs",
	MEMapEditorMode_Interactables: "Interactables",
}

type MEMapEditor struct {
	worldFile          game.GameWorldFile
	camera             *MECamera
	tileDrawer         *METileDrawer
	npcDrawer          *MENpcDrawer
	interactableDrawer *MEInteractableDrawer
	ui                 *MEUi

	mode MEMapEditorMode

	lastErrorMessage string
	lastErrorTime    time.Time

	worldWidth, worldHeight int
}

func NewMEMapEditor() (*MEMapEditor, error) {
	worldFile, err := game.LoadGameWorldFile(config.GameWorldFilePath)
	if err != nil {
		return nil, err
	}

	worldHeight := len(worldFile.Tiles)
	worldWidth := len(worldFile.Tiles[0])

	mapEditor := &MEMapEditor{
		worldWidth:  worldWidth,
		worldHeight: worldHeight,
	}
	getMapEditor := func() *MEMapEditor {
		return mapEditor
	}

	camera := NewMECamera(0, 0, 1, 5)
	getCamera := func() *MECamera {
		return camera
	}

	tileDrawer := NewMETileDrawer(getCamera)
	getTileDrawer := func() *METileDrawer {
		return tileDrawer
	}

	npcDrawer := NewMENpcDrawer(getCamera, getMapEditor)
	getNpcDrawer := func() *MENpcDrawer {
		return npcDrawer
	}
	npcDrawer.Init(worldFile.Npcs)

	interactableDrawer := NewMEInteractableDrawer(getCamera, getMapEditor)
	getInteractableDrawer := func() *MEInteractableDrawer {
		return interactableDrawer
	}
	interactableDrawer.Init(worldFile.Interactables)

	ui := NewMEUi(getTileDrawer, getNpcDrawer, getInteractableDrawer, getMapEditor)

	mapEditor.worldFile = worldFile
	mapEditor.camera = camera
	mapEditor.tileDrawer = tileDrawer
	mapEditor.ui = ui
	mapEditor.npcDrawer = npcDrawer
	mapEditor.interactableDrawer = interactableDrawer
	return mapEditor, nil
}

func (m *MEMapEditor) Update() error {

	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		m.mode = MEMapEditorMode((int(m.mode) + 1) % mapEditorModeCount)
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.lastErrorMessage = m.Save()
		m.lastErrorTime = time.Now()
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		m.Resize(m.worldWidth+1, m.worldHeight+1)
		m.worldWidth++
		m.worldHeight++
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyShift) && inpututil.IsKeyJustPressed(ebiten.KeyEqual) {
		m.Resize(m.worldWidth+10, m.worldHeight+10)
		m.worldWidth += 10
		m.worldHeight += 10
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) && inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		m.Resize(m.worldWidth-1, m.worldHeight-1)
		m.worldWidth--
		m.worldHeight--
	}

	if ebiten.IsKeyPressed(ebiten.KeyControl) && ebiten.IsKeyPressed(ebiten.KeyShift) && inpututil.IsKeyJustPressed(ebiten.KeyMinus) {
		m.Resize(m.worldWidth-10, m.worldHeight-10)
		m.worldWidth -= 10
		m.worldHeight -= 10
	}

	m.camera.Update()

	switch m.mode {
	case MEMapEditorMode_Tiles:
		m.tileDrawer.Update(m.worldFile.Tiles)

	case MEMapEditorMode_Npcs:
		m.npcDrawer.Update(m.worldFile.Npcs)

	case MEMapEditorMode_Interactables:
		m.interactableDrawer.Update(m.worldFile.Interactables)
	}

	return nil
}

func (m *MEMapEditor) Draw(screen *ebiten.Image) {
	m.tileDrawer.Draw(screen, m.worldFile.Tiles)
	m.npcDrawer.Draw(screen, m.worldFile.Npcs, m.mode)
	m.interactableDrawer.Draw(screen, m.worldFile.Interactables, m.mode)
	m.ui.Draw(screen, m.mode)
}

func (m *MEMapEditor) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 1280, 720
}

func (m *MEMapEditor) Save() string {
	data, err := json.MarshalIndent(m.worldFile, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to marshal world file: %s", err)
	}

	err = os.WriteFile(config.GameWorldFilePath, data, 0644)
	if err != nil {
		return fmt.Sprintf("failed to write world file: %s", err)
	}

	return "Saved!"
}

func (m *MEMapEditor) Resize(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("width and height must be greater than 0")
	}

	newTiles := make([][]int, height)

	for y := range newTiles {
		newTiles[y] = make([]int, width)
	}

	oldHeight := len(m.worldFile.Tiles)
	if oldHeight == 0 {
		m.worldFile.Tiles = newTiles
		return nil
	}

	oldWidth := len(m.worldFile.Tiles[0])

	copyHeight := min(oldHeight, height)
	copyWidth := min(oldWidth, width)

	for y := 0; y < copyHeight; y++ {
		copy(newTiles[y][:copyWidth], m.worldFile.Tiles[y][:copyWidth])
	}

	m.worldFile.Tiles = newTiles

	return nil
}
