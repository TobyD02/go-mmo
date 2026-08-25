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

type MEInteractableDrawer struct {
	getCam               func() *MECamera
	activeInteractable   int
	brushSize            int
	spawnPos             *util.Vec2
	interactableIDs      []string
	interactableRegistry game.GInteractableRegistry
	isErasing            bool

	worldWidth  int
	worldHeight int

	interactableColors map[string]color.Color
	spatialIndex       util.GSpatialIndex
}

func NewMEInteractableDrawer(getCam func() *MECamera, worldWidth, worldHeight int) *MEInteractableDrawer {
	interactableRegistry, err := game.GetInteractableRegistry()
	if err != nil {
		panic(err)
	}

	interactableIDs := make([]string, 0, len(interactableRegistry))
	for id := range interactableRegistry {
		interactableIDs = append(interactableIDs, id)
	}

	interactableColors := make(map[string]color.Color)
	for _, id := range interactableIDs {
		hash := sha256.Sum256([]byte(id))

		interactableColors[id] = color.RGBA{
			R: hash[0],
			G: hash[1],
			B: hash[2],
			A: 255,
		}
	}

	return &MEInteractableDrawer{
		getCam:               getCam,
		activeInteractable:   0,
		spawnPos:             nil,
		brushSize:            1,
		interactableIDs:      interactableIDs,
		interactableRegistry: interactableRegistry,

		worldWidth:  worldWidth,
		worldHeight: worldHeight,

		interactableColors: interactableColors,
		spatialIndex:       util.NewGSpatialIndex(),
	}
}

func (t *MEInteractableDrawer) Init(interactables map[string][]util.Vec2) {
	for id, positions := range interactables {
		for _, pos := range positions {
			t.spatialIndex.Add(id, pos)
		}
	}
}

func (t *MEInteractableDrawer) getActiveInteractableName() string {
	return t.interactableIDs[t.activeInteractable]
}

func (t *MEInteractableDrawer) Draw(screen *ebiten.Image, interactables map[string][]util.Vec2, activeMode MEMapEditorMode) {

	isActive := activeMode == MEMapEditorMode_Interactables

	for idx, interactableID := range t.interactableIDs {
		if idx == t.activeInteractable {
			continue
		}
		col, exists := t.interactableColors[interactableID]
		if !exists {
			panic("missing color for interactable: " + interactableID)
		}
		for _, pos := range interactables[interactableID] {
			realX, realY := t.getCam().WorldToScreenPos(float32(pos.X), float32(pos.Y))
			realSize := t.getCam().WorldToScreenScale(METileSize)
			vector.FillRect(screen, realX, realY, realSize, realSize, col, false)

			if isActive && t.interactableIDs[t.activeInteractable] == interactableID {
				vector.StrokeRect(screen, realX, realY, realSize, realSize, 1, color.White, false)
			}
		}
	}
	activeInteractableID := t.interactableIDs[t.activeInteractable]
	col, exists := t.interactableColors[activeInteractableID]
	if !exists {
		panic("missing color for interactable: " + activeInteractableID)
	}
	for _, pos := range interactables[activeInteractableID] {
		realX, realY := t.getCam().WorldToScreenPos(float32(pos.X), float32(pos.Y))
		realSize := t.getCam().WorldToScreenScale(METileSize)
		vector.FillRect(screen, realX, realY, realSize, realSize, col, false)

		if isActive && t.interactableIDs[t.activeInteractable] == activeInteractableID {
			vector.StrokeRect(screen, realX, realY, realSize, realSize, 1, color.White, false)
		}
	}
}

func (t *MEInteractableDrawer) Update(interactables map[string][]util.Vec2) {
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		t.activeInteractable = (t.activeInteractable + 1) % len(t.interactableIDs)
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
	interactableX := int(x) / METileSize
	interactableY := int(y) / METileSize

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		t.forEachBrushTile(interactableX, interactableY, func(x, y int) {
			if t.interactableSafe(x, y) {
				t.setInteractable(x, y, interactables)
			}
		})
	}

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		t.forEachBrushTile(interactableX, interactableY, func(x, y int) {
			if t.positionInBounds(x, y) {
				t.eraseInteractable(x, y, interactables)
			}
		})
	}
}

func (t *MEInteractableDrawer) forEachBrushTile(
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

func (t *MEInteractableDrawer) setInteractable(interactableX, interactableY int, interactables map[string][]util.Vec2) {

	interactableID := t.interactableIDs[t.activeInteractable]
	interactables[interactableID] = append(interactables[interactableID], util.Vec2{X: interactableX, Y: interactableY})

	t.spatialIndex.Add(interactableID, util.Vec2{X: interactableX, Y: interactableY})
}

func (t *MEInteractableDrawer) eraseInteractable(x, y int, interactables map[string][]util.Vec2) {
	pos := util.Vec2{X: x, Y: y}

	ids := t.spatialIndex.QueryPos(x, y)

	for id := range ids {
		positions := interactables[id]

		for i, entityPos := range positions {
			if entityPos.Equal(pos) {

				interactables[id] = append(
					positions[:i],
					positions[i+1:]...,
				)

				t.spatialIndex.Remove(id, pos)

				return
			}
		}
	}
}

func (t *MEInteractableDrawer) positionInBounds(x, y int) bool {
	return x >= 0 &&
		x < t.worldWidth &&
		y >= 0 &&
		y < t.worldHeight
}

func (t *MEInteractableDrawer) interactableSafe(x, y int) bool {
	if !t.positionInBounds(x, y) {
		return false
	}

	return len(t.spatialIndex.QueryPos(x, y)) == 0
}
