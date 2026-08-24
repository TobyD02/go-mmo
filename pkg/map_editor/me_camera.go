package map_editor

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type MECamera struct {
	X, Y  float32
	Zoom  float32
	Speed float32
}

func NewMECamera(x, y, zoom, speed float32) *MECamera {
	return &MECamera{
		X:     x,
		Y:     y,
		Zoom:  zoom,
		Speed: speed,
	}
}

func (c *MECamera) Update() {
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		c.X -= c.Speed / c.Zoom
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		c.X += c.Speed / c.Zoom
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		c.Y -= c.Speed / c.Zoom
	}

	if ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		c.Y += c.Speed / c.Zoom
	}

	_, scrollY := ebiten.Wheel()
	if scrollY > 0 {
		c.Zoom *= 1.1
	} else if scrollY < 0 {
		c.Zoom /= 1.1
	}
}

func (c *MECamera) GetMousePosition() (float32, float32) {
	mouseX, mouseY := ebiten.CursorPosition()
	x := float32(mouseX)/c.Zoom + c.X
	y := float32(mouseY)/c.Zoom + c.Y

	return x, y
}

func (c *MECamera) Draw(screen *ebiten.Image, col color.Color, x, y int) {

	size := float32(METileSize) * c.Zoom
	screenX := (float32(x*METileSize) - c.X) * c.Zoom
	screenY := (float32(y*METileSize) - c.Y) * c.Zoom
	vector.FillRect(screen, screenX, screenY, size, size, col, false)
}

func (c *MECamera) WorldToScreenPos(x, y float32) (float32, float32) {
	screenX := (float32(x*METileSize) - c.X) * c.Zoom
	screenY := (float32(y*METileSize) - c.Y) * c.Zoom

	return screenX, screenY
}

func (c *MECamera) WorldToScreenScale(size int) float32 {
	return float32(size) * c.Zoom
}
