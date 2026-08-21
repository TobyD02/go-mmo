package game

import (
	"github.com/google/uuid"
)

type GInteractable struct {
	Name                string     `json:"name"`
	ID                  string     `json:"id"`
	MaxTickCooldown     int        `json:"max_tick_cooldown"`
	TickWorkForYield    int        `json:"tick_work_for_yield"`
	MaxOccupantCooldown int        `json:"max_occupant_cooldown"`
	LootPool            *GLootPool `json:"loot_pool"`
}

type GInteractableInstance struct {
	ID                  string
	InteractableID      string
	Pos                 Vec2
	CurrentTickCooldown int
	CurrentTicksWorked  int
	OccupiedBy          string
	LastTickWorked      int
	OccupantCooldown    int
}

func NewGInteractableInstance(interactableID string, x, y int) *GInteractableInstance {
	return &GInteractableInstance{
		ID:                  uuid.NewString(),
		InteractableID:      interactableID,
		Pos:                 Vec2{X: x, Y: y},
		CurrentTickCooldown: 0,
		CurrentTicksWorked:  0,
		OccupiedBy:          "",
		LastTickWorked:      0,
		OccupantCooldown:    0,
	}
}

func (i *GInteractableInstance) DoWork(currentTick int) {
	i.CurrentTicksWorked++
	i.LastTickWorked = currentTick
}

func (i *GInteractableInstance) DidWorkThisTick(currentTick int) bool {
	return i.LastTickWorked == currentTick
}

func (i *GInteractableInstance) ClearOccupant() {
	i.OccupiedBy = ""
}

func (i *GInteractableInstance) IsOccupied() bool {
	return i.OccupiedBy != ""
}

func (i *GInteractableInstance) WorkIsDone() bool {
	interactable := GetInteractableFromRegistry(i.InteractableID)
	if interactable == nil {
		return false
	}

	return i.CurrentTicksWorked >= interactable.TickWorkForYield
}

func (i *GInteractableInstance) GetYieldAndTriggerCooldown() map[string]int {
	interactable := GetInteractableFromRegistry(i.InteractableID)
	if interactable == nil {
		return map[string]int{}
	}

	i.CurrentTickCooldown = interactable.MaxTickCooldown
	return interactable.LootPool.GetYield()
}

// PlayerCanOccupyOrWork - returns true if the player meets the conditions for working the Interactable.
// this includes being within range, also taking occupancy of the interactable
func (i *GInteractableInstance) PlayerCanOccupyOrWork(player *GPlayer) bool {
	interactable := GetInteractableFromRegistry(i.InteractableID)
	if interactable == nil {
		return false
	}

	if player.Pos.DistanceSquared(i.Pos) > 2 { // distance squared == 2 allows diagonals
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

	i.OccupantCooldown = interactable.MaxOccupantCooldown // Reset occupant cooldown once assured of occupancy
	return true
}
