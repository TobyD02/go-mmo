// Package game - global module for game logic
package game

type GameWorldTile int

const (
	TileBlank GameWorldTile = iota
	TilePlayer
)

type GameWorld struct {
	Map      [][]GameWorldTile
	Entities map[string]*GPosition
}

type GPosition struct {
	x, y int
}

func NewGameWorld(width, height int) *GameWorld {
	gameMap := make([][]GameWorldTile, height)
	for y := range gameMap {
		gameMap[y] = make([]GameWorldTile, width)
	}

	gameEntities := make(map[string]*GPosition)

	return &GameWorld{
		Map:      gameMap,
		Entities: gameEntities,
	}
}

func (g *GameWorld) AddPlayer(playerID string, x, y int) {
	g.Map[y][x] = TilePlayer
	g.Entities[playerID] = &GPosition{x: x, y: y}
}

func (g *GameWorld) MovePlayer(playerID string, x, y int) {
	player := g.Entities[playerID]

	if (player.y+y < 0 || player.y+y >= len(g.Map)) ||
		(player.x+x < 0 || player.x+x >= len(g.Map[player.y+y])) {
		return
	}

	g.Map[player.y][player.x] = TileBlank

	player.x += x
	player.y += y

	g.Map[player.y][player.x] = TilePlayer
}
