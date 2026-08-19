package game

type GInventoryItem struct {
	Item     *GItem `json:"item"`
	Quantity int    `json:"quantity"`
}
