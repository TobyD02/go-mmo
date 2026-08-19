package game

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
	LastTickWorked      int    `json:"last_tick_worked"`
	OccupantCooldown    int    `json:"occupant_cooldown"`
	MaxOccupantCooldown int    `json:"max_occupant_cooldown"`
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
		LastTickWorked:      0,
		OccupantCooldown:    0,
		MaxOccupantCooldown: 3,
	}
}

func (i *GInteractable) DoWork(currentTick int) {
	i.CurrentTicksWorked++
	i.LastTickWorked = currentTick
}

func (i *GInteractable) DidWorkThisTick(currentTick int) bool {
	return i.LastTickWorked == currentTick
}

func (i *GInteractable) ClearOccupant() {
	i.OccupiedBy = ""
}

func (i *GInteractable) IsOccupied() bool {
	return i.OccupiedBy != ""
}

func (i *GInteractable) WorkIsDone() bool {
	return i.CurrentTicksWorked >= i.TickWorkForYield
}

func (i *GInteractable) GetYieldAndTriggerCooldown() (*GItem, int) {
	i.CurrentTickCooldown = i.MaxTickCooldown
	yieldAmount := rand.IntN(i.YieldAmountMax-i.YieldAmountMin) + i.YieldAmountMin
	return &i.Yield, yieldAmount
}

// PlayerCanOccupyOrWork - returns true if the player meets the conditions for working the Interactable.
// this includes being within range, also taking occupancy of the interactable
func (i *GInteractable) PlayerCanOccupyOrWork(player *GPlayer) bool {
	if player.Pos.Distance(i.Pos) > 1 {
		return false // Cannot
	}

	if i.CurrentTickCooldown != 0 {
		return false // Interactable is on cooldown
	}

	if i.OccupiedBy == "" {
		i.OccupiedBy = player.ID
	} else if i.OccupiedBy != player.ID {
		return false
	}

	i.OccupantCooldown = i.MaxOccupantCooldown // Reset occupant cooldown once assured of occupancy
	return true
}
