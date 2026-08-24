package game

type GameWorldTile int

const (
	TileFloor GameWorldTile = iota
	TileWall
	TileSpawn
)

func CanWalk(t GameWorldTile) bool {
	if t == TileWall {
		return false
	}

	return true
}
