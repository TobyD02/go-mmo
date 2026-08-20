package game

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type GInteractableRegistry map[string]*GInteractable

var (
	interactableRegistry     GInteractableRegistry
	interactableRegistryOnce sync.Once
)

func GetInteractableRegistry() (GInteractableRegistry, error) {
	var initErr error

	interactableRegistryOnce.Do(func() {
		initErr = initInteractableRegistry()
	})

	if initErr != nil {
		return nil, initErr
	}

	return interactableRegistry, nil
}

func initInteractableRegistry() error {
	interactableConfig := os.Getenv("INTERACTABLES_CONFIG")
	if interactableConfig == "" {
		// default to a local dir for now
		// return fmt.Errorf("no interactable config file path received")
		interactableConfig = "./data/interactables.json"
	}

	data, err := os.ReadFile(interactableConfig)
	if err != nil {
		return fmt.Errorf("failed to read interactable config file: %w", err)
	}

	registry := make(GInteractableRegistry)

	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to decode interactable config: %w", err)
	}

	// Set ID's
	for id, interactable := range registry {
		interactable.ID = id
		registry[id] = interactable
	}

	interactableRegistry = registry

	return nil
}

func GetInteractableFromRegistry(interactableID string) *GInteractable {
	interactableRegistry, err := GetInteractableRegistry()
	if err != nil {
		return nil
	}

	interactable, exists := interactableRegistry[interactableID]
	if !exists {
		return nil
	}

	return interactable
}

func GetInteractableNameFromRegistry(interactableID string) string {
	interactableRegistry, err := GetInteractableRegistry()
	if err != nil {
		return "N/A"
	}

	interactable, exists := interactableRegistry[interactableID]
	if !exists {
		return "N/A"
	}

	return interactable.Name
}
