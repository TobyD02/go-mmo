package game_common

import "github.com/google/uuid"

type GPlayer struct {
	ID        string            `json:"id"`
	Pos       Vec2              `json:"pos"`
	Inventory []*GInventoryItem `json:"inventory"`
}

func NewGPlayer(x, y int) *GPlayer {

	inventory := make([]*GInventoryItem, 0)
	return &GPlayer{
		ID:        uuid.NewString(),
		Pos:       Vec2{X: x, Y: y},
		Inventory: inventory,
	}
}

func (p *GPlayer) AddToInventory(item *GItem, quantity int) {
	for _, inventoryItem := range p.Inventory {
		if inventoryItem.Item.ID == item.ID {
			inventoryItem.Quantity += quantity
			return
		}
	}

	p.Inventory = append(p.Inventory, &GInventoryItem{Item: item, Quantity: quantity})
}
