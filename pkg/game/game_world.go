// Package game - game logic shared between server and client
package game

type GameWorld struct {
	Map           [][]GameWorldTile
	Players       map[string]*GPlayer
	Interactables map[string]*GInteractable
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
	gameInteractables := make(map[string]*GInteractable)

	return &GameWorld{
		Map:           gameMap,
		Players:       gamePlayers,
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

	// Interactables
	if g.Interactables == nil {
		g.Interactables = make(map[string]*GInteractable)
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

func (g *GameWorld) QueryInteractableAtPosition(x, y int) *GInteractable { // singular since there can only be one
	for _, i := range g.Interactables {
		if i.Pos.X == x && i.Pos.Y == y {
			return i
		}
	}

	return nil
}
