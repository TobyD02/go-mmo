package game

import (
	"math/rand/v2"

	"github.com/google/uuid"
)

type GNpc struct {
	Name               string    `json:"name"`
	ID                 string    `json:"id"`
	LootPool           GLootPool `json:"loot_pool"` // itemID: amount
	Aggressive         bool      `json:"aggressive"`
	WillFlee           bool      `json:"will_flee"`
	TickThinkFrequency int       `json:"think_tick_frequency"`
}

type GNpcInstance struct {
	ID              string `json:"id"`
	NpcID           string `json:"npc_id"`
	Pos             Vec2   `json:"pos"`
	LastPos         Vec2   `json:"last_pos"`
	LastTickUpdated int    `json:"last_tick_updated"`

	PlayerTargetID string `json:"player_target_id"` // could give them multiple targets - maybe just prioritise player?
	NpcTargetID    string `json:"npc_target_id"`    // player should take priority
}

func NewGNpcInstance(npcID string, x, y int) *GNpcInstance {
	return &GNpcInstance{
		ID:      uuid.NewString(),
		NpcID:   npcID,
		Pos:     Vec2{X: x, Y: y},
		LastPos: Vec2{X: x, Y: y},

		PlayerTargetID: "",
		NpcTargetID:    "",
	}
}

func (n *GNpcInstance) GetLoot() map[string]int {
	npc := GetNpcFromRegistry(n.NpcID)
	if npc == nil {
		return map[string]int{}
	}

	return npc.LootPool.GetYield()
}

// Think - returns a move delta
func (n *GNpcInstance) Think(currentTick int) Vec2 {
	// so - if no target, it should patrol

	n.LastTickUpdated = currentTick
	if n.PlayerTargetID == "" && n.NpcTargetID == "" {
		delta := n.patrolDecision()
		return delta
	}

	// if player target - check distance
	// if within range, attack
	// else move closer
	// if npc target

	return Vec2{X: 0, Y: 0}

}

func (n *GNpcInstance) patrolDecision() Vec2 {
	// just pick a random direction for now. keep it limited to 1 of either axis
	moveX := rand.Float32() > 0.5
	direction := rand.IntN(2)*2 - 1

	if moveX {
		return Vec2{X: direction, Y: 0}
	} else {
		return Vec2{X: 0, Y: direction}
	}
}
