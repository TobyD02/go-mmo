package game

type GPlayer struct {
	ID        string         `json:"id"`
	Pos       Vec2           `json:"pos"`
	Inventory map[string]int `json:"inventory"` // itemID: amount
}

func NewGPlayer(id string, x, y int) *GPlayer {

	inventory := make(map[string]int)
	return &GPlayer{
		ID:        id,
		Pos:       Vec2{X: x, Y: y},
		Inventory: inventory,
	}
}

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
