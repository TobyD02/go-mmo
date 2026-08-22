package util

import (
	"math"
)

type Vec2 struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func (v Vec2) DistanceSquared(v2 Vec2) float64 {
	xDiff := v2.X - v.X
	yDiff := v2.Y - v.Y

	return math.Pow(float64(xDiff), 2) + math.Pow(float64(yDiff), 2)
}

func (v Vec2) Distance(v2 Vec2) float64 {
	return math.Sqrt(v.DistanceSquared(v2))
}

func (v Vec2) LengthSquared() float64 {
	return math.Pow(float64(v.X), 2) + math.Pow(float64(v.Y), 2)
}

func (v Vec2) Length() float64 {
	return math.Sqrt(float64(v.LengthSquared()))
}

func (v Vec2) Direction(v2 Vec2) Vec2 {
	return Vec2{
		X: sign(v2.X - v.X),
		Y: sign(v2.Y - v.Y),
	}
}

func (v Vec2) Normalize() Vec2 {
	return Vec2{
		X: sign(v.X),
		Y: sign(v.Y),
	}
}

func (v Vec2) Reverse() Vec2 {
	return Vec2{
		X: -v.X,
		Y: -v.Y,
	}
}

func (v Vec2) Equal(v2 Vec2) bool {
	return v.X == v2.X && v.Y == v2.Y
}

func sign(n int) int {
	switch {
	case n < 0:
		return -1
	case n > 0:
		return 1
	default:
		return 0
	}
}
