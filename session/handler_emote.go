package session

import (
	"github.com/df-mc/dragonfly/server/event"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// EmoteHandler is a handler that handles emote packets.
type EmoteHandler struct{}

func (h *EmoteHandler) Handle(ctx *event.Context, p packet.Packet, s *Session) error {
	s.c.(interface{ SendSettingsForm() }).SendSettingsForm()
	return nil
}
