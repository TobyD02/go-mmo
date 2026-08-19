package game

import "github.com/google/uuid"

type GItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func NewGItem(name string) GItem {
	return GItem{
		ID:   uuid.NewString(),
		Name: name,
	}
}

// @TODO - get rid of this
var TestItem GItem = NewGItem("someItem")
