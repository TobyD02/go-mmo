package game

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type GNpcRegistry map[string]*GNpc

var (
	npcRegistry     GNpcRegistry
	npcRegistryOnce sync.Once
)

func GetNpcRegistry() (GNpcRegistry, error) {
	var initErr error

	npcRegistryOnce.Do(func() {
		initErr = initNpcRegistry()
	})

	if initErr != nil {
		return nil, initErr
	}

	return npcRegistry, nil
}

func initNpcRegistry() error {
	npcConfig := os.Getenv("NPCS_CONFIG")
	if npcConfig == "" {
		// default to a local dir for now
		// return fmt.Errorf("no npc config file path received")
		npcConfig = "./data/npcs.json"
	}

	data, err := os.ReadFile(npcConfig)
	if err != nil {
		return fmt.Errorf("failed to read npc config file: %w", err)
	}

	registry := make(GNpcRegistry)

	if err := json.Unmarshal(data, &registry); err != nil {
		return fmt.Errorf("failed to decode npc config: %w", err)
	}

	// Set ID's
	for id, npc := range registry {
		npc.ID = id
		registry[id] = npc
	}

	npcRegistry = registry

	return nil
}

func GetNpcFromRegistry(npcID string) *GNpc {
	npcRegistry, err := GetNpcRegistry()
	if err != nil {
		return nil
	}

	npc, exists := npcRegistry[npcID]
	if !exists {
		return nil
	}

	return npc
}

func GetNpcNameFromRegistry(npcID string) string {
	npcRegistry, err := GetNpcRegistry()
	if err != nil {
		return "N/A"
	}

	npc, exists := npcRegistry[npcID]
	if !exists {
		return "N/A"
	}

	return npc.Name
}
