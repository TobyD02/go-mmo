package client

import (
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/util"
)

// Will have all private - and then client side renderer uses query methods
type GClientWorld struct {
	gameWorld                        *game.GameWorld
	playerSpatialIndex               util.GSpatialIndex
	npcInstanceSpatialIndex          util.GSpatialIndex
	interactableInstanceSpatialIndex util.GSpatialIndex
}

func NewGClientWorld(gameWorld *game.GameWorld) *GClientWorld {
	c := &GClientWorld{
		gameWorld:                        gameWorld,
		playerSpatialIndex:               make(util.GSpatialIndex),
		npcInstanceSpatialIndex:          make(util.GSpatialIndex),
		interactableInstanceSpatialIndex: make(util.GSpatialIndex),
	}

	// Initialise spatial indexes

	for id, player := range gameWorld.Players {
		c.playerSpatialIndex.Add(id, player.GetPos())
	}

	for id, npc := range gameWorld.Npcs {
		c.npcInstanceSpatialIndex.Add(id, npc.GetPos())
	}

	for id, interactable := range gameWorld.Interactables {
		c.interactableInstanceSpatialIndex.Add(id, interactable.GetPos())
	}

	return c
}

func (c *GClientWorld) ApplyWorldDiff(diff *game.GameWorldDiff) {
	updateSpatialIndex(
		c.npcInstanceSpatialIndex,
		c.gameWorld.Npcs,
		diff.NpcsDiff,
	)

	updateSpatialIndex(
		c.playerSpatialIndex,
		c.gameWorld.Players,
		diff.PlayersDiff,
	)

	updateSpatialIndex(
		c.interactableInstanceSpatialIndex,
		c.gameWorld.Interactables,
		diff.InteractablesDiff,
	)

	c.gameWorld.ApplyDiff(diff)
}

func updateSpatialIndex[E any, T interface {
	*E
	game.GEntity
}](index util.GSpatialIndex, world map[string]T, diff map[string]T) {
	for id, newEntity := range diff {
		oldEntity := world[id]

		if newEntity == nil {
			if oldEntity != nil {
				index.Remove(id, oldEntity.GetPos())
			}

			continue
		}

		if oldEntity != nil {
			index.Update(
				id,
				oldEntity.GetPos(),
				newEntity.GetPos(),
			)
		} else {
			index.Add(id, newEntity.GetPos())
		}
	}
}

func (c *GClientWorld) QueryPlayer(id string) *game.GPlayer {
	return c.gameWorld.Players[id]
}

func (c *GClientWorld) QueryPlayersAtPosition(x, y int) map[string]*game.GPlayer {
	ids := c.playerSpatialIndex.QueryPos(x, y)

	players := make(map[string]*game.GPlayer, len(ids))

	for id := range ids {
		if player := c.gameWorld.Players[id]; player != nil {
			players[id] = player
		}
	}

	return players
}

func (c *GClientWorld) QueryNpcInstancesAtPosition(x, y int) map[string]*game.GNpcInstance {
	ids := c.npcInstanceSpatialIndex.QueryPos(x, y)

	npcs := make(map[string]*game.GNpcInstance, len(ids))

	for id := range ids {
		if npc := c.gameWorld.Npcs[id]; npc != nil {
			npcs[id] = npc
		}
	}

	return npcs
}

func (c *GClientWorld) QueryInteractableInstanceAtPosition(x, y int) *game.GInteractableInstance {
	ids := c.interactableInstanceSpatialIndex.QueryPos(x, y)

	for id := range ids {
		return c.gameWorld.Interactables[id]
	}

	return nil
}

func (c *GClientWorld) QueryMap(x, y int) game.GameWorldTile {
	return c.gameWorld.QueryMap(x, y)
}

func (c *GClientWorld) IsInBounds(x, y int) bool {
	return x >= 0 &&
		x < c.gameWorld.Width &&
		y >= 0 &&
		y < c.gameWorld.Height
}

func (c *GClientWorld) Width() int {
	return c.gameWorld.Width
}

func (c *GClientWorld) Height() int {
	return c.gameWorld.Height
}
