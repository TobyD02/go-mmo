package main

import (
	"image/color"
	"log"
	"math/rand"
	"os"

	"github.com/google/uuid"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/tobyd02/go-mmo/pkg/client"
	"github.com/tobyd02/go-mmo/pkg/game"
)

const tileSize = 16

type RenderPosition struct {
	x float64
	y float64
}

type Game struct {
	gameWorld *game.GameWorld
	client    *client.GClient

	renderPositions map[string]RenderPosition
}

func NewGame(serverURI, clientID string) (*Game, error) {
	c := client.NewGClient(false)

	world, err := c.Start(serverURI, clientID)
	if err != nil {
		return nil, err
	}

	g := &Game{
		gameWorld:       world,
		client:          c,
		renderPositions: make(map[string]RenderPosition),
	}

	for id, player := range world.Players {
		g.renderPositions[id] = RenderPosition{
			x: float64(player.Pos.X * tileSize),
			y: float64(player.Pos.Y * tileSize),
		}
	}

	return g, nil
}

func (g *Game) Update() error {
	g.client.Update()

	moveX := rand.Intn(3) - 1
	moveY := rand.Intn(3) - 1

	if moveX != 0 || moveY != 0 {
		log.Printf("SEND MOVE: x=%d y=%d", moveX, moveY)
		g.client.SendMoveMessage(moveX, moveY)
	}

	diff, err := g.client.ReadGameWorldDiff()
	if err != nil {
		return err
	}

	if diff != nil {
		log.Printf("RECEIVED WORLD DIFF")
		g.gameWorld.ApplyDiff(diff)
	}

	// Lerp every connected player toward their server position.
	for id, player := range g.gameWorld.Players {
		targetX := float64(player.Pos.X * tileSize)
		targetY := float64(player.Pos.Y * tileSize)

		render, exists := g.renderPositions[id]
		if !exists {
			render = RenderPosition{
				x: targetX,
				y: targetY,
			}
		}

		render.x += (targetX - render.x) * 0.15
		render.y += (targetY - render.y) * 0.15

		g.renderPositions[id] = render
	}

	// Remove render positions for disconnected players.
	for id := range g.renderPositions {
		if _, exists := g.gameWorld.Players[id]; !exists {
			delete(g.renderPositions, id)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for id := range g.gameWorld.Players {
		render := g.renderPositions[id]

		ebitenutil.DrawRect(
			screen,
			render.x,
			render.y,
			tileSize,
			tileSize,
			color.White,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 800, 600
}

func (g *Game) Close() {
	g.client.StopAndCloseConnection()
}

func main() {
	serverURI := os.Getenv("G_SERVER")
	if serverURI == "" {
		serverURI = "ws://localhost:8080"
	}

	clientID := uuid.NewString()

	log.Printf("Connecting to %s", serverURI)

	game, err := NewGame(serverURI, clientID)
	if err != nil {
		log.Fatal(err)
	}
	defer game.Close()

	log.Printf(
		"Connected! ClientID=%s Players=%d",
		game.client.ClientID,
		len(game.gameWorld.Players),
	)

	ebiten.SetWindowSize(800, 600)
	ebiten.SetWindowTitle("Go MMO")

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
