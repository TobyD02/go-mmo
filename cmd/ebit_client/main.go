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

type Game struct {
	gameWorld *game.GameWorld
	client    *client.GClient
}

func NewGame(serverURI, clientID string) (*Game, error) {
	c := client.NewGClient(false)

	world, err := c.Start(serverURI, clientID)
	if err != nil {
		return nil, err
	}

	return &Game{
		gameWorld: world,
		client:    c,
	}, nil
}

func (g *Game) Update() error {
	// Process incoming server data.
	g.client.Update()

	// Send a random movement every Ebitengine tick.
	moveX := rand.Intn(3) - 1
	moveY := rand.Intn(3) - 1

	if moveX != 0 || moveY != 0 {
		log.Printf("SEND MOVE: x=%d y=%d", moveX, moveY)
		g.client.SendMoveMessage(moveX, moveY)
	}

	// Read incoming world changes.
	diff, err := g.client.ReadGameWorldDiff()
	if err != nil {
		return err
	}

	if diff != nil {
		log.Printf("RECEIVED WORLD DIFF")
		g.gameWorld.ApplyDiff(diff)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for _, player := range g.gameWorld.Players {
		x := float64(player.Pos.X * tileSize)
		y := float64(player.Pos.Y * tileSize)

		ebitenutil.DrawRect(
			screen,
			x,
			y,
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
