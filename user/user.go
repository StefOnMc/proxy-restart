package user

import (
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/mymaqc/proxy/entity"
	"github.com/mymaqc/proxy/session"
	"golang.org/x/oauth2"
)

// User is a user of the server.
type User struct {
	*entity.Entity
	settings Settings
	s        *session.Session
	pos      atomic.Value[mgl32.Vec3]
	rot      atomic.Value[mgl32.Vec2]
}

// New creates a new user for the given session.
func New(s *session.Session) *User {
	u := &User{
		s:      s,
		Entity: entity.New(s.GameData().EntityRuntimeID),
	}
	return u
}

// Name returns the name of the user.
func (u *User) Name() string {
	return u.s.Name()
}

// SendForm sends a form to the user.
func (u *User) SendForm(f form.Form) {
	u.s.SendForm(f)
}

func (u *User) SetSettings(s Settings) {
	u.settings = s
}

func (u *User) Settings() Settings {
	return u.settings
}

func (u *User) Transfer(addr string, src *TokenSource) {
	u.s.Transfer(addr, src)
}
func (u *User) WaitForTransfer() (string, oauth2.TokenSource) {
	return u.s.WaitForTransfer()
}
