package game

type GameWorldDiff struct {
	MapDiff           []GameWorldMapDiff                `json:"map_diff"`
	PlayersDiff       map[string]*GPlayer               `json:"players_diff"`
	InteractablesDiff map[string]*GInteractableInstance `json:"interactables_diff"`
}

type GameWorldMapDiff struct {
	Pos  Vec2          `json:"position"`
	Tile GameWorldTile `json:"tile"`
}

func (g *GameWorldDiff) IsEmpty() bool {
	return len(g.MapDiff) == 0 && len(g.PlayersDiff) == 0 && len(g.InteractablesDiff) == 0
}
