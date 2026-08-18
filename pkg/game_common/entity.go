package game

type GTag int

const (
	GPlayer GTag = iota
)

type GEntity struct {
	ID   string  `json:"id"`
	Pos  Vec2    `json:"pos"`
	Tags *[]GTag `json:"tags"`
}
