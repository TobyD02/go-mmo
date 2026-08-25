package map_editor

import (
	"crypto/sha256"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/util"
)

type MENpcDrawer struct {
	getCam      func() *MECamera
	activeNpc   int
	brushSize   int
	spawnPos    *util.Vec2
	npcIDs      []string
	npcRegistry game.GNpcRegistry
	isErasing   bool

	getMapEditor func() *MEMapEditor

	npcColors    map[string]color.Color
	spatialIndex util.GSpatialIndex
}

func NewMENpcDrawer(getCam func() *MECamera, getMapEditor func() *MEMapEditor) *MENpcDrawer {
	npcRegistry, err := game.GetNpcRegistry()
	if err != nil {
		panic(err)
	}

	npcIDs := make([]string, 0, len(npcRegistry))
	for id := range npcRegistry {
		npcIDs = append(npcIDs, id)
	}

	npcColors := make(map[string]color.Color)
	for _, id := range npcIDs {
		hash := sha256.Sum256([]byte(id))

		npcColors[id] = color.RGBA{
			R: hash[0],
			G: hash[1],
			B: hash[2],
			A: 255,
		}
	}

	return &MENpcDrawer{
		getCam:      getCam,
		activeNpc:   0,
		spawnPos:    nil,
		brushSize:   1,
		npcIDs:      npcIDs,
		npcRegistry: npcRegistry,

		getMapEditor: getMapEditor,

		npcColors:    npcColors,
		spatialIndex: util.NewGSpatialIndex(),
	}
}

func (t *MENpcDrawer) Init(npcs map[string][]util.Vec2) {
	for id, positions := range npcs {
		for _, pos := range positions {
			t.spatialIndex.Add(id, pos)
		}
	}
}

func (t *MENpcDrawer) getActiveNpcName() string {
	return t.npcIDs[t.activeNpc]
}

func (t *MENpcDrawer) Draw(screen *ebiten.Image, npcs map[string][]util.Vec2, activeMode MEMapEditorMode) {

	isActive := activeMode == MEMapEditorMode_Npcs

	for idx, npcID := range t.npcIDs {
		if idx == t.activeNpc {
			continue
		}
		col, exists := t.npcColors[npcID]
		if !exists {
			panic("missing color for npc: " + npcID)
		}
		for _, pos := range npcs[npcID] {
			realX, realY := t.getCam().WorldToScreenPos(float32(pos.X), float32(pos.Y))
			realSize := t.getCam().WorldToScreenScale(METileSize)
			vector.FillRect(screen, realX, realY, realSize, realSize, col, false)

			if isActive && t.npcIDs[t.activeNpc] == npcID {
				vector.StrokeRect(screen, realX, realY, realSize, realSize, 1, color.White, false)
			}
		}
	}
	activeNpcID := t.npcIDs[t.activeNpc]
	col, exists := t.npcColors[activeNpcID]
	if !exists {
		panic("missing color for npc: " + activeNpcID)
	}
	for _, pos := range npcs[activeNpcID] {
		realX, realY := t.getCam().WorldToScreenPos(float32(pos.X), float32(pos.Y))
		realSize := t.getCam().WorldToScreenScale(METileSize)
		vector.FillRect(screen, realX, realY, realSize, realSize, col, false)

		if isActive && t.npcIDs[t.activeNpc] == activeNpcID {
			vector.StrokeRect(screen, realX, realY, realSize, realSize, 1, color.White, false)
		}
	}
}

func (t *MENpcDrawer) Update(npcs map[string][]util.Vec2) {
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		t.activeNpc = (t.activeNpc + 1) % len(t.npcIDs)
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		t.brushSize++
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
		t.brushSize--
		if t.brushSize < 1 {
			t.brushSize = 1
		}
	}

	x, y := t.getCam().GetMousePosition()
	npcX := int(x) / METileSize
	npcY := int(y) / METileSize

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		t.forEachBrushTile(npcX, npcY, func(x, y int) {
			if t.npcSafe(x, y) {
				t.setNpc(x, y, npcs)
			}
		})
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		t.forEachBrushTile(npcX, npcY, func(x, y int) {
			if t.positionInBounds(x, y) {
				t.eraseNpc(x, y, npcs)
			}
		})
	}
}

func (t *MENpcDrawer) forEachBrushTile(
	centerX, centerY int,
	fn func(x, y int),
) {
	startX := centerX - t.brushSize/2
	startY := centerY - t.brushSize/2

	for y := startY; y < startY+t.brushSize; y++ {
		for x := startX; x < startX+t.brushSize; x++ {
			fn(x, y)
		}
	}
}

func (t *MENpcDrawer) setNpc(npcX, npcY int, npcs map[string][]util.Vec2) {

	npcID := t.npcIDs[t.activeNpc]
	npcs[npcID] = append(npcs[npcID], util.Vec2{X: npcX, Y: npcY})

	t.spatialIndex.Add(npcID, util.Vec2{X: npcX, Y: npcY})
}

func (t *MENpcDrawer) eraseNpc(x, y int, npcs map[string][]util.Vec2) {
	pos := util.Vec2{X: x, Y: y}

	ids := t.spatialIndex.QueryPos(x, y)

	for id := range ids {
		positions := npcs[id]

		for i, entityPos := range positions {
			if entityPos.Equal(pos) {

				npcs[id] = append(
					positions[:i],
					positions[i+1:]...,
				)

				t.spatialIndex.Remove(id, pos)

				return
			}
		}
	}
}

func (t *MENpcDrawer) positionInBounds(x, y int) bool {
	return x >= 0 &&
		x < t.getMapEditor().worldWidth &&
		y >= 0 &&
		y < t.getMapEditor().worldHeight
}

func (t *MENpcDrawer) npcSafe(x, y int) bool {
	if !t.positionInBounds(x, y) {
		return false
	}

	return len(t.spatialIndex.QueryPos(x, y)) == 0
}
