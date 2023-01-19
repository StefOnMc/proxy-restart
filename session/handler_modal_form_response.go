package session

import (
	"bytes"
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync"
)

// ModalFormResponseHandler handles the ModalFormResponse packet.
type ModalFormResponseHandler struct {
	mu        sync.Mutex
	forms     map[uint32]form.Form
	currentID atomic.Uint32
}

// nullBytes contains the word 'null' converted to a byte slice.
var nullBytes = []byte("null\n")

// Handle ...
func (h *ModalFormResponseHandler) Handle(ctx *event.Context, p packet.Packet, s *Session) error {
	pk := p.(*packet.ModalFormResponse)
	h.mu.Lock()
	f, ok := h.forms[pk.FormID]
	delete(h.forms, pk.FormID)
	h.mu.Unlock()
	rsp, _ := pk.ResponseData.Value()
	if !ok && bytes.Equal(rsp, nullBytes) {
		// Sometimes the client seems to send a second response with "null" as the response, which would
		// cause the player to be kicked by the server. This should patch that.
		return nil
	}
	if bytes.Equal(rsp, nullBytes) || len(rsp) == 0 {
		// The form was cancelled: The cross in the top right corner was clicked.
		rsp = nil
	}
	if !ok {
		return fmt.Errorf("no form with ID %v currently opened", pk.FormID)
	}
	if err := f.SubmitJSON(rsp, s.c); err != nil {
		return fmt.Errorf("error submitting form data: %w", err)
	}
	return nil
}
