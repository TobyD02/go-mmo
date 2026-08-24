package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tobyd02/go-mmo/pkg/map_editor"
)

func main() {
	mapEditor, err := map_editor.NewMEMapEditor()
	if err != nil {
		log.Fatal(err)
	}

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("Map Editor")

	if err := ebiten.RunGame(mapEditor); err != nil {
		log.Fatal(err)
	}
}
