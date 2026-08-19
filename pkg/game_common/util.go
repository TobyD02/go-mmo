package game_common

import "math"

type Vec2 struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (v Vec2) Distance(v2 Vec2) float32 {
	xDiff := v2.X - v.X
	yDiff := v2.Y - v.Y

	return float32(
		math.Sqrt(
			math.Pow(float64(xDiff), 2) - math.Pow(float64(yDiff), 2),
		),
	)
}
