package game

import (
	"math/rand/v2"

	"github.com/tobyd02/go-mmo/pkg/util"
)

type GPlayer struct {
	ID        string         `json:"id"`
	Pos       util.Vec2      `json:"pos"`
	Inventory map[string]int `json:"inventory"` // itemID: amount
}

func NewGPlayer(id string, x, y int) *GPlayer {

	inventory := make(map[string]int)
	return &GPlayer{
		ID:        id,
		Pos:       util.Vec2{X: x, Y: y},
		Inventory: inventory,
	}
}

// Implement GEntity methods

func (p *GPlayer) GetID() string {
	return p.ID
}

func (p *GPlayer) GetPos() util.Vec2 {
	return p.Pos
}

// -------------------------

func (p *GPlayer) AddToInventory(itemsAndAmount map[string]int) {

	for itemID, amount := range itemsAndAmount {
		_, exists := p.Inventory[itemID]
		if !exists {
			p.Inventory[itemID] = amount
		} else {
			p.Inventory[itemID] = p.Inventory[itemID] + amount
		}

	}
}

func (p *GPlayer) DoDamage() int {
	return rand.IntN(4) + 1
}
