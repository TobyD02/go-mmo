package game

import "github.com/tobyd02/go-mmo/pkg/util"

type GEntity interface {
	GetID() string
	GetPos() util.Vec2
}
