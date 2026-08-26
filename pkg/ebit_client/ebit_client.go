package ebit_client

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tobyd02/go-mmo/pkg/client"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/util"
)

// TODO: Client side prediction
// TODO: Position Interpolation
// TODO: Fix tile sizing/ not drawing to full screen
// TODO: Assets

// TODO: Mouse click = queue move to that position. - each tick, pop a move message and send the next.
// TODO: On a new mouse click, replace queued messages with new path.

type GEbitClient struct {
	client   *client.GClient
	tileSize int

	windowWidth  int
	windowHeight int

	viewportWidth  int
	viewportHeight int

	mousePath      []util.Vec2
	mousePathIndex int
}

func NewGEbitClient(client *client.GClient) (*GEbitClient, error) {

	//viewportWidth, viewportHeight := config.ClientSimulationRangeX*2, config.ClientSimulationRangeY*2
	windowWidth, windowHeight := 1280, 720

	tileSize := 10

	viewportWidth, viewportHeight := windowWidth/tileSize, windowHeight/tileSize

	//tileSize := min(
	//	windowWidth/viewportWidth,
	//	windowHeight/viewportHeight,
	//)

	return &GEbitClient{
		client:       client,
		tileSize:     tileSize,
		windowWidth:  windowWidth,
		windowHeight: windowHeight,

		viewportWidth:  viewportWidth,
		viewportHeight: viewportHeight,
	}, nil
}

func (c *GEbitClient) Update() error {
	if err := c.client.Update(); err != nil {
		return err
	}

	mouseX, mouseY := ebiten.CursorPosition()
	self := c.client.QuerySelf()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		t := c.screenToWorld(mouseX, mouseY)
		c.mousePath = c.client.GetPathTo(t)
		c.mousePathIndex = 0
	}

	if c.mousePathIndex < len(c.mousePath)-1 {
		if self.Pos.Equal(c.mousePath[c.mousePathIndex]) {
			c.mousePathIndex++
		}

		move := c.client.GetMovesTo(c.mousePath[c.mousePathIndex])[0]
		_ = c.client.Move(move.X, move.Y)

	}

	return nil
}

func (c *GEbitClient) Draw(screen *ebiten.Image) {

	clientPlayer := c.client.QuerySelf()
	if clientPlayer == nil {
		ebitenutil.DebugPrint(screen, "Loading")
		return
	}

	centerX := clientPlayer.Pos.X
	centerY := clientPlayer.Pos.Y

	startX := centerX - c.viewportWidth/2
	startY := centerY - c.viewportHeight/2

	endX := startX + c.viewportWidth
	endY := startY + c.viewportHeight

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			if !c.client.IsInBounds(x, y) {
				continue
			}

			screenX := x - startX
			screenY := y - startY

			players := c.client.QueryPlayers(x, y)
			interactableInstance := c.client.QueryInteractable(x, y)
			npcInstances := c.client.QueryNpcs(x, y)

			if len(players) > 0 {
				if players[c.client.ClientID] != nil {
					c.drawRect(screen, screenX, screenY, color.RGBA{R: 255, G: 255, B: 255, A: 255})
				} else {
					c.drawRect(screen, screenX, screenY, color.RGBA{R: 255, G: 255, B: 0, A: 255})
				}

				continue
			}

			if interactableInstance != nil {
				c.drawRect(screen, screenX, screenY, color.RGBA{R: 0, G: 255, B: 0, A: 255})
				continue
			}

			if len(npcInstances) > 0 {
				for _ = range npcInstances {
					c.drawRect(screen, screenX, screenY, color.RGBA{R: 0, G: 0, B: 255, A: 255})
					break
				}
				continue
			}

			tile := c.client.QueryTile(x, y)
			switch tile {
			case game.TileWall:
				c.drawRect(screen, screenX, screenY, color.RGBA{R: 100, G: 100, B: 100, A: 255})
			case game.TileSpawn:
				c.drawRect(screen, screenX, screenY, color.RGBA{R: 100, G: 100, B: 240, A: 255})
			default:
			}
		}
	}

	for _, pos := range c.mousePath {
		screenX := pos.X - startX
		screenY := pos.Y - startY

		c.drawRectStroke(
			screen,
			screenX,
			screenY,
			color.RGBA{R: 150, G: 150, B: 150, A: 200},
		)

	}

}
func (c *GEbitClient) Layout(outsideWidth, outsideHeight int) (int, int) {
	return c.windowWidth, c.windowHeight
}

func (c *GEbitClient) Close() {
	c.client.StopAndCloseConnection()
}

func (c *GEbitClient) drawRect(screen *ebiten.Image, x, y int, col color.Color) {
	vector.FillRect(screen, float32(x*c.tileSize), float32(y*c.tileSize), float32(c.tileSize), float32(c.tileSize), col, false)
}

func (c *GEbitClient) drawRectStroke(screen *ebiten.Image, x, y int, col color.Color) {
	vector.StrokeRect(screen, float32(x*c.tileSize), float32(y*c.tileSize), float32(c.tileSize), float32(c.tileSize), 1, col, false)
}

func (c *GEbitClient) screenToWorld(
	screenX, screenY int,
) util.Vec2 {
	clientPlayer := c.client.QuerySelf()
	if clientPlayer == nil {
		return util.Vec2{}
	}

	startX := clientPlayer.Pos.X - c.viewportWidth/2
	startY := clientPlayer.Pos.Y - c.viewportHeight/2

	return util.Vec2{
		X: startX + screenX/c.tileSize,
		Y: startY + screenY/c.tileSize,
	}
}
