package game

import (
	"github.com/google/uuid"
)

type GInteractable struct {
	ID                  string `json:"id"`
	Pos                 Vec2   `json:"pos"`
	MaxTickCooldown     int    `json:"max_tick_cooldown"`
	CurrentTickCooldown int    `json:"current_tick_cooldown"`
	TickWorkForYield    int    `json:"tick_work_for_yield"`
	CurrentTicksWorked  int    `json:"current_ticks_worked"`
	OccupiedBy          string `json:"occupied_by"` // entitiy ID?
	LastTickWorked      int    `json:"last_tick_worked"`
	OccupantCooldown    int    `json:"occupant_cooldown"`
	MaxOccupantCooldown int    `json:"max_occupant_cooldown"`

	LootPool *GLootPool `json:"loot_pool"`
}

func NewGInteractable(x, y int, lootPool *GLootPool) *GInteractable {
	return &GInteractable{
		ID:                  uuid.NewString(),
		Pos:                 Vec2{X: x, Y: y},
		MaxTickCooldown:     15,
		CurrentTickCooldown: 0,
		TickWorkForYield:    2,
		CurrentTicksWorked:  0,
		OccupiedBy:          "",
		LastTickWorked:      0,
		OccupantCooldown:    0,
		MaxOccupantCooldown: 3,

		LootPool: lootPool,
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

func (i *GInteractable) GetYieldAndTriggerCooldown() map[string]int {
	i.CurrentTickCooldown = i.MaxTickCooldown
	return i.LootPool.GetYield()
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
