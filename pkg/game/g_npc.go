package game

import (
	"log"
	"math/rand/v2"

	"github.com/google/uuid"
)

type GNpc struct {
	Name        string    `json:"name"`
	ID          string    `json:"id"`
	LootPool    GLootPool `json:"loot_pool"` // itemID: amount
	Aggressive  bool      `json:"aggressive"`
	WillFlee    bool      `json:"will_flee"`
	PatrolSpeed int       `json:"patrol_speed"`
	CombatSpeed int       `json:"combat_speed"`
	MaxHealth   int       `json:"max_health"`
}

type GNpcInstance struct {
	ID              string `json:"id"`
	NpcID           string `json:"npc_id"`
	Pos             Vec2   `json:"pos"`
	LastPos         Vec2   `json:"last_pos"`
	LastTickUpdated int    `json:"last_tick_updated"`
	Health          int    `json:"health"`

	healthHasSet bool

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
		healthHasSet:   false,
		Health:         0,
	}
}

func (n *GNpcInstance) GetLoot() map[string]int {
	npc := GetNpcFromRegistry(n.NpcID)
	if npc == nil {
		return map[string]int{}
	}

	return npc.LootPool.GetYield()
}

func (n *GNpcInstance) checkInit() {
	if !n.healthHasSet {
		npc := GetNpcFromRegistry(n.NpcID)
		if npc != nil {
			n.Health = npc.MaxHealth
			n.healthHasSet = true
		}
	}
}

// Think - returns a move delta
func (n *GNpcInstance) Think(currentTick int, playerTarget *GPlayer, npcTarget *GNpcInstance) Vec2 {
	n.checkInit()

	// so - if no target, it should patrol
	n.LastTickUpdated = currentTick

	if playerTarget == nil && n.PlayerTargetID != "" {
		n.PlayerTargetID = ""
	}

	if npcTarget == nil && n.NpcTargetID != "" {
		n.NpcTargetID = ""
	}

	var moveDecision Vec2
	if playerTarget != nil {
		moveDecision = n.handlePlayerTarget(playerTarget)
	} else if npcTarget != nil {
		moveDecision = n.handleNpcTarget(npcTarget)
	} else {
		moveDecision = n.patrolDecision()
	}

	n.LastPos = n.Pos

	// Only move on 1 axis - if moving on both, randomly set 1 axis to 0
	if moveDecision.LengthSquared() > 1 {
		if rand.Float32() > 0.5 {
			moveDecision.Y = 0
		} else {
			moveDecision.X = 0
		}
	}

	return moveDecision
}

func (n *GNpcInstance) handlePlayerTarget(playerTarget *GPlayer) Vec2 {
	return n.fleeDecision(playerTarget.Pos)
}

func (n *GNpcInstance) handleNpcTarget(npcTarget *GNpcInstance) Vec2 {
	return n.fleeDecision(npcTarget.Pos)
}

func (n *GNpcInstance) fleeDecision(targetPos Vec2) Vec2 {

	if n.Pos.Distance(targetPos) == 0 || n.Pos.Equal(n.LastPos) {
		return n.patrolDecision()
	}
	return n.Pos.Direction(targetPos).Reverse().Normalize()
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

func (n *GNpcInstance) PlayerCanAttack(player *GPlayer) bool {
	n.checkInit()

	return n.canAttack(player.Pos)
}

func (n *GNpcInstance) canAttack(pos Vec2) bool {
	npc := GetNpcFromRegistry(n.NpcID)
	if npc == nil {
		log.Println("cannot find npc")
		return false
	}

	if pos.Distance(n.Pos) > 1 {
		log.Println("too far")
		return false // Cannot
	}

	if n.Health <= 0 {
		log.Println("health is <= 0")
		return false // Npc is on cooldown
	}

	return true
}

func (n *GNpcInstance) TakePlayerDamage(player *GPlayer) int {
	dmg := player.DoDamage()
	n.Health -= dmg
	if n.Health <= 0 {
		n.Health = 0
	}

	if n.PlayerTargetID == "" {
		n.PlayerTargetID = player.ID
	}

	return dmg
}

func (n *GNpcInstance) CanDoCombat(currentServerTick int) bool {
	npc := GetNpcFromRegistry(n.NpcID)

	return n.HasTarget() && currentServerTick-n.LastTickUpdated >= npc.CombatSpeed
}

func (n *GNpcInstance) CanDoPatrol(currentServerTick int) bool {
	npc := GetNpcFromRegistry(n.NpcID)

	return !n.HasTarget() && currentServerTick-n.LastTickUpdated >= npc.PatrolSpeed
}

func (n *GNpcInstance) HasTarget() bool {
	return n.PlayerTargetID != "" || n.NpcTargetID != ""
}
