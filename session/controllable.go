package session

import (
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/go-gl/mathgl/mgl32"
)

type Controllable interface {
	SendForm(form.Form)
	Name() string
	Position() mgl32.Vec3
	SetPosition(mgl32.Vec3)
	Rotate(mgl32.Vec2)
	Rotation() mgl32.Vec2
}
