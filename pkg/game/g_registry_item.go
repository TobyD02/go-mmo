package game

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type GItemRegistry map[string]*GItem

var (
	itemRegistry     GItemRegistry
	itemRegistryOnce sync.Once
)

func GetItemRegistry() (GItemRegistry, error) {
	var initErr error

	itemRegistryOnce.Do(func() {
		initErr = initItemRegistry()
	})

	if initErr != nil {
		return nil, initErr
	}

	return itemRegistry, nil
}

func initItemRegistry() error {
	itemConfig := os.Getenv("ITEM_CONFIG")
	if itemConfig == "" {
		// default to a local dir for now
		// return fmt.Errorf("no item config file path received")
		itemConfig = "./data/items.json"
	}

	data, err := os.ReadFile(itemConfig)
	if err != nil {
		return fmt.Errorf("failed to read item config file: %w", err)
	}

	registry := make(GItemRegistry)

	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to decode item config: %w", err)
	}

	// Set ID's
	for id, item := range registry {
		item.ID = id
		registry[id] = item
	}

	itemRegistry = registry

	return nil
}

func GetItemNameFromRegistry(itemID string) string {
	itemRegistry, err := GetItemRegistry()
	if err != nil {
		return "N/A"
	}

	item, exists := itemRegistry[itemID]
	if !exists {
		return "N/A"
	}

	return item.Name
}
