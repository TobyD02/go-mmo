package client

import (
	"github.com/tobyd02/go-mmo/pkg/config"
	"github.com/tobyd02/go-mmo/pkg/game"
	"github.com/tobyd02/go-mmo/pkg/util"

	"github.com/matteo00gm/go-astar"
)

// Will have all private - and then client side renderer uses query methods
type GClientWorld struct {
	gameWorld                        *game.GameWorld
	playerSpatialIndex               util.GSpatialIndex
	npcInstanceSpatialIndex          util.GSpatialIndex
	interactableInstanceSpatialIndex util.GSpatialIndex

	pathFinder       *astar.Astar
	pathFinderOrigin util.Vec2
}

func NewGClientWorld(gameWorld *game.GameWorld) *GClientWorld {
	c := &GClientWorld{
		gameWorld:                        gameWorld,
		playerSpatialIndex:               util.NewGSpatialIndex(),
		npcInstanceSpatialIndex:          util.NewGSpatialIndex(),
		interactableInstanceSpatialIndex: util.NewGSpatialIndex(),
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

func (c *GClientWorld) PredictClientMovement(dx, dy int, clientID string) {
	if dx == 0 && dy == 0 {
		return
	}

	player, exists := c.gameWorld.Players[clientID]
	if !exists { // Safe guard
		return
	}

	newX := player.Pos.X + dx
	newY := player.Pos.Y + dy

	if (player.Pos.Y+dy < 0 || player.Pos.Y+dy >= len(c.gameWorld.Map)) ||
		(player.Pos.X+dx < 0 || player.Pos.X+dx >= len(c.gameWorld.Map[player.Pos.Y+dy])) {
		return
	}

	if !game.CanWalk(c.gameWorld.QueryMap(newX, newY)) {
		return // Cannot move to unwalkable tile
	}

	if len(c.interactableInstanceSpatialIndex.QueryPos(newX, newY)) > 0 {
		return // Cannot move over interactables?
	}

	playerOldPos := player.Pos

	player.Pos.X += dx
	player.Pos.Y += dy

	c.playerSpatialIndex.Update(clientID, playerOldPos, player.Pos)
}

func (w *GClientWorld) BuildPathFinder(
	center util.Vec2,
) {
	origin := util.Vec2{
		X: center.X - config.ClientSimulationRangeX/2,
		Y: center.Y - config.ClientSimulationRangeY/2,
	}

	w.pathFinderOrigin = origin

	grid := make([][]int, config.ClientSimulationRangeY)

	for localY := range config.ClientSimulationRangeY {
		grid[localY] = make([]int, config.ClientSimulationRangeX)

		for localX := range config.ClientSimulationRangeX {
			worldPos := util.Vec2{
				X: origin.X + localX,
				Y: origin.Y + localY,
			}

			tile := w.QueryMap(worldPos.X, worldPos.Y)
			interactable := w.interactableInstanceSpatialIndex.QueryPos(worldPos.X, worldPos.Y)

			if !game.CanWalk(tile) || len(interactable) > 0 {
				grid[localY][localX] = 1
			}
		}
	}

	w.pathFinder = astar.New(
		grid,
		&astar.EuclideanHeuristic{},
	)
}

func (w *GClientWorld) GetPath(
	start util.Vec2,
	target util.Vec2,
) []util.Vec2 {

	if !w.isInPathFinderBounds(start) || !w.isInPathFinderBounds(target) {
		return nil
	}

	startCoords := astar.Coords{
		X: start.X - w.pathFinderOrigin.X,
		Y: start.Y - w.pathFinderOrigin.Y,
	}

	targetCoords := astar.Coords{
		X: target.X - w.pathFinderOrigin.X,
		Y: target.Y - w.pathFinderOrigin.Y,
	}

	found, path := w.pathFinder.FindPath(
		startCoords,
		targetCoords,
	)

	if !found {
		return nil
	}

	worldPath := make([]util.Vec2, len(path))

	for i, pos := range path {
		worldPath[i] = util.Vec2{
			X: pos.X + w.pathFinderOrigin.X,
			Y: pos.Y + w.pathFinderOrigin.Y,
		}
	}

	return worldPath
}

func (w *GClientWorld) isInPathFinderBounds(
	pos util.Vec2,
) bool {
	return pos.X >= w.pathFinderOrigin.X &&
		pos.X < w.pathFinderOrigin.X+config.ClientSimulationRangeX &&
		pos.Y >= w.pathFinderOrigin.Y &&
		pos.Y < w.pathFinderOrigin.Y+config.ClientSimulationRangeY
}

func (w *GClientWorld) PathToMoves(
	path []util.Vec2,
) []util.Vec2 {
	if len(path) < 2 {
		return nil
	}

	moves := make(
		[]util.Vec2,
		0,
		len(path)-1,
	)

	for i := 1; i < len(path); i++ {
		moves = append(moves, util.Vec2{
			X: path[i].X - path[i-1].X,
			Y: path[i].Y - path[i-1].Y,
		})
	}

	return moves
}
