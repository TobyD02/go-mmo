package map_editor

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/util"
)

var tileColorMap = map[game.GameWorldTile]METileData{
	game.TileFloor: {
		Name: "Floor",
		Col:  color.RGBA{R: 33, G: 33, B: 33, A: 255},
	},
	game.TileWall: {
		Name: "Wall",
		Col:  color.RGBA{R: 233, G: 233, B: 233, A: 255},
	},
	game.TileSpawn: {
		Name: "Spawn",
		Col:  color.RGBA{R: 255, G: 255, B: 0, A: 255},
	},
}

func getTileName(tile game.GameWorldTile) string {
	return tileColorMap[game.GameWorldTile(tile)].Name
}

func getTileCol(tile game.GameWorldTile) color.Color {
	return tileColorMap[game.GameWorldTile(tile)].Col
}

type METileData struct {
	Name string
	Col  color.Color
}

type METileDrawer struct {
	getCam     func() *MECamera
	activeTile game.GameWorldTile
	brushSize  int
	spawnPos   *util.Vec2
}

func NewMETileDrawer(getCam func() *MECamera) *METileDrawer {
	return &METileDrawer{
		getCam:     getCam,
		activeTile: game.TileWall,
		spawnPos:   nil,
		brushSize:  1,
	}
}

func (t *METileDrawer) Draw(screen *ebiten.Image, tileMap [][]int) {

	for y, row := range tileMap {
		for x, tile := range row {
			realX, realY := t.getCam().WorldToScreenPos(float32(x), float32(y))
			realSize := t.getCam().WorldToScreenScale(METileSize)
			vector.FillRect(screen, realX, realY, realSize, realSize, getTileCol(game.GameWorldTile(tile)), false)
		}
	}

}

func (t *METileDrawer) Update(tileMap [][]int) {
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		t.activeTile = game.TileFloor
	}

	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		t.activeTile = game.TileWall
	}

	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		t.activeTile = game.TileSpawn
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

	if t.spawnPos == nil {
		for y, row := range tileMap {
			for x, tile := range row {
				if game.GameWorldTile(tile) == game.TileSpawn {
					t.spawnPos = &util.Vec2{X: x, Y: y}
					break
				}
			}
		}
	}

	x, y := t.getCam().GetMousePosition()
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {

		if t.activeTile == game.TileSpawn && t.spawnPos != nil {
			return
		}

		tileX := int(x) / METileSize
		tileY := int(y) / METileSize

		if t.brushSize == 1 {
			if t.tileSafe(tileX, tileY, tileMap) {
				t.setTile(tileX, tileY, tileMap)
			}
			return
		}

		for brushY := -t.brushSize / 2; brushY < t.brushSize/2; brushY++ {
			for brushX := -t.brushSize / 2; brushX < t.brushSize/2; brushX++ {

				targetX := tileX + brushX
				targetY := tileY + brushY

				if t.tileSafe(targetX, targetY, tileMap) {
					t.setTile(targetX, targetY, tileMap)
				}
			}
		}

	}
}

func (t *METileDrawer) setTile(tileX, tileY int, tileMap [][]int) {
	tileMap[tileY][tileX] = int(t.activeTile)

	pos := util.Vec2{
		X: tileX,
		Y: tileY,
	}

	// Creating the spawn
	if t.activeTile == game.TileSpawn {
		t.spawnPos = &pos
		return
	}

	// Overwriting the existing spawn
	if t.spawnPos != nil && t.spawnPos.Equal(pos) {
		t.spawnPos = nil
	}
}

func (t *METileDrawer) tileSafe(tileX, tileY int, tileMap [][]int) bool {
	if t.activeTile == game.TileSpawn && t.spawnPos != nil {
		return false
	}

	if tileY < 0 || tileY >= len(tileMap) {
		return false
	}

	if tileX < 0 || tileX >= len(tileMap[tileY]) {
		return false
	}

	if t.activeTile != game.TileSpawn && t.activeTile != game.TileFloor {
		if t.spawnPos != nil &&
			!(t.spawnPos.Distance(util.Vec2{X: tileX, Y: tileY}) > 10) {
			return false
		}
	}

	return true
}
