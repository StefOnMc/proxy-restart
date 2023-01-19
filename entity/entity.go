package entity

import (
	"github.com/df-mc/atomic"
	"github.com/go-gl/mathgl/mgl32"
)

type Entity struct {
	runtimeID uint64
	pos       atomic.Value[mgl32.Vec3]
	rot       atomic.Value[mgl32.Vec2]
}

func New(runtimeID uint64) *Entity {
	return &Entity{
		runtimeID: runtimeID,
	}
}

func (e *Entity) RuntimeID() uint64 {
	return e.runtimeID
}

func (e *Entity) Position() mgl32.Vec3 {
	return e.pos.Load()
}

func (e *Entity) SetPosition(pos mgl32.Vec3) {
	e.pos.Store(pos)
}

func (e *Entity) Rotation() mgl32.Vec2 {
	return e.rot.Load()
}

func (e *Entity) Rotate(rot mgl32.Vec2) {
	e.rot.Store(rot)
}
