// Package game - game logic shared between server and client
package game

type GameWorld struct {
	Map           [][]GameWorldTile
	Players       map[string]*GPlayer
	Npcs          map[string]*GNpcInstance
	Interactables map[string]*GInteractableInstance
	Width         int
	Height        int
	SpawnPoint    Vec2
}

func NewGameWorld(width, height int) *GameWorld {
	gameMap := make([][]GameWorldTile, height)
	for y := range gameMap {
		gameMap[y] = make([]GameWorldTile, width)
	}

	gamePlayers := make(map[string]*GPlayer)
	gameNpcs := make(map[string]*GNpcInstance)
	gameInteractables := make(map[string]*GInteractableInstance)

	return &GameWorld{
		Map:           gameMap,
		Players:       gamePlayers,
		Npcs:          gameNpcs,
		Interactables: gameInteractables,
		Width:         width, Height: height,
		SpawnPoint: Vec2{int(width / 2), int(height / 2)},
	}
}

func (g *GameWorld) ApplyDiff(diff *GameWorldDiff) {
	for _, mapDiff := range diff.MapDiff {
		pos := mapDiff.Pos
		tile := mapDiff.Tile

		g.Map[pos.Y][pos.X] = tile
	}

	// Players
	if g.Players == nil {
		g.Players = make(map[string]*GPlayer)
	}

	for id, player := range diff.PlayersDiff {
		if player == nil {
			delete(g.Players, id)
		} else {
			g.Players[id] = player
		}
	}

	// Npcs
	if g.Npcs == nil {
		g.Npcs = make(map[string]*GNpcInstance)
	}

	for id, npc := range diff.NpcsDiff {
		if npc == nil {
			delete(g.Npcs, id)
		} else {
			g.Npcs[id] = npc
		}
	}

	// Interactables
	if g.Interactables == nil {
		g.Interactables = make(map[string]*GInteractableInstance)
	}

	for id, interactable := range diff.InteractablesDiff {
		if interactable == nil {
			delete(g.Interactables, id)
		} else {
			g.Interactables[id] = interactable
		}
	}
}

func (g *GameWorld) QueryMap(x, y int) GameWorldTile {
	if x <= 0 || x >= g.Width || y <= 0 || y >= g.Height {
		return TileWall
	}

	return g.Map[y][x]
}

func (g *GameWorld) QueryPlayersAtPosition(x, y int) map[string]*GPlayer {
	players := make(map[string]*GPlayer, 0)
	for playerID, e := range g.Players {
		if e.Pos.X == x && e.Pos.Y == y {
			players[playerID] = e
		}
	}

	return players
}

func (g *GameWorld) QueryInteractableInstanceAtPosition(x, y int) *GInteractableInstance { // singular since there can only be one
	for _, i := range g.Interactables {
		if i.Pos.X == x && i.Pos.Y == y {
			return i
		}
	}

	return nil
}

func (g *GameWorld) QueryNpcInstanceAtPosition(x, y int) *GNpcInstance { // singular since there can only be one
	for _, i := range g.Npcs {
		if i.Pos.X == x && i.Pos.Y == y {
			return i
		}
	}

	return nil
}
