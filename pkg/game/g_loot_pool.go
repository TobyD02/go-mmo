package game

import (
	"math/rand/v2"
)

type GLootPool []GLoot

type GLoot struct {
	ItemID         string  `json:"item_id"`
	YieldAmountMin int     `json:"yield_amount_min"`
	YieldAmountMax int     `json:"yield_amount_max"`
	YieldChance    float32 `json:"yield_chance"`
}

// / NewGLootPool- item ids
func NewGLootPool(itemIDs []string) *GLootPool {
	lootPool := make(GLootPool, 0, len(itemIDs))
	for _, itemID := range itemIDs {

		lootPool = append(lootPool, GLoot{
			ItemID:         itemID,
			YieldAmountMin: 1,
			YieldAmountMax: 3,
			YieldChance:    0.5,
		})
	}

	return &lootPool
}

// GetYield - returns ["itemID": amount]
func (l *GLootPool) GetYield() map[string]int {
	yields := make(map[string]int, 0)
	for _, loot := range *l {
		if rand.Float32() > loot.YieldChance {
			continue
		}

		yieldAmount := loot.YieldAmountMin
		if loot.YieldAmountMin != loot.YieldAmountMax {
			// rand.IntN isn't inclusive on the upper end - so + 1
			yieldAmount += rand.IntN(loot.YieldAmountMax - loot.YieldAmountMin + 1)
		}

		yields[loot.ItemID] = yieldAmount
	}

	return yields
}
