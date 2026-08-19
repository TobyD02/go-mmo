// Package game_common - Common game logic shared between client and server
package game_common

import (
	"math/rand/v2"

	"github.com/google/uuid"
)

type GInteractable struct {
	ID                  string `json:"id"`
	Pos                 Vec2   `json:"pos"`
	Yield               GItem  `json:"yield"`
	YieldAmountMin      int    `json:"yield_amount_min"`
	YieldAmountMax      int    `json:"yield_amount_max"`
	MaxTickCooldown     int    `json:"max_tick_cooldown"`
	CurrentTickCooldown int    `json:"current_tick_cooldown"`
	TickWorkForYield    int    `json:"tick_work_for_yield"`
	CurrentTicksWorked  int    `json:"current_ticks_worked"`
	OccupiedBy          string `json:"occupied_by"` // entitiy ID?
}

func NewGInteractable(x, y int, yield GItem) *GInteractable {
	return &GInteractable{
		ID:                  uuid.NewString(),
		Pos:                 Vec2{X: x, Y: y},
		Yield:               yield,
		YieldAmountMin:      1,
		YieldAmountMax:      rand.IntN(3) + 2,
		MaxTickCooldown:     15,
		CurrentTickCooldown: 0,
		TickWorkForYield:    2,
		CurrentTicksWorked:  0,
		OccupiedBy:          "",
	}
}
