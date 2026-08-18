package game

type GameWorldTile int

const (
	TileWalkable GameWorldTile = iota
	TileWall
)

var TileChars = map[GameWorldTile]string{
	TileWalkable: "..",
	TileWall:     "██",
}
