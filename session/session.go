package session

import (
	"encoding/json"
	"fmt"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/player/form"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"golang.org/x/oauth2"
	"time"
)

// Session is a session for a client.
type Session struct {
	c Controllable

	remoteConn *minecraft.Conn
	conn       *minecraft.Conn

	handlers map[uint32]packetHandler

	gameData minecraft.GameData

	waitForTransfer chan transferData
}

type transferData struct {
	addr string
	ts   oauth2.TokenSource
}

// New creates a new session for the given client.
func New(conn *minecraft.Conn, gameData minecraft.GameData) *Session {
	s := &Session{
		gameData:        gameData,
		conn:            conn,
		handlers:        make(map[uint32]packetHandler),
		waitForTransfer: make(chan transferData),
	}
	s.registerHandlers()
	go s.handlePackets()
	go s.sendMovement()
	return s
}

// GameData returns the game data of the session.
func (s *Session) GameData() minecraft.GameData {
	return s.gameData
}

func (s *Session) Transfer(addr string, src oauth2.TokenSource) {
	s.waitForTransfer <- transferData{addr: addr, ts: src}
	close(s.waitForTransfer)
}

func (s *Session) WaitForTransfer() (string, oauth2.TokenSource) {
	d := <-s.waitForTransfer
	return d.addr, d.ts
}

// Name returns the name of the session.
func (s *Session) Name() string {
	return s.conn.IdentityData().DisplayName
}

// SetControllable sets the controllable for the session.
func (s *Session) SetControllable(c Controllable) {
	s.c = c
}

// SetRemoteConn sets the remote connection for the session.
func (s *Session) SetRemoteConn(c *minecraft.Conn) {
	s.remoteConn = c
}

// SendForm sends a form to the client.
func (s *Session) SendForm(f form.Form) {
	b, _ := json.Marshal(f)

	h := s.handlers[packet.IDModalFormResponse].(*ModalFormResponseHandler)
	id := h.currentID.Add(69)

	h.mu.Lock()
	if len(h.forms) > 10 {
		//fmt.printf("SendForm %v: more than 10 active forms: dropping an existing one.\n", s.c.Name())
		for k := range h.forms {
			delete(h.forms, k)
			break
		}
	}
	h.forms[id] = f
	h.mu.Unlock()
	s.writePacket(&packet.ModalFormRequest{
		FormID:   id,
		FormData: b,
	})
}

// writePacket writes a packet to the client.
func (s *Session) writePacket(p packet.Packet) {
	err := s.conn.WritePacket(p)
	if err != nil {
		fmt.Printf("failed writing packet to %v (%v): %v\n", s.conn.RemoteAddr(), s.c.Name(), err)
	}
}

// handlePacket handles an incoming packet, processing it accordingly. If the packet had invalid data or was
// otherwise not valid in its context, an error is returned.
func (s *Session) handlePacket(ctx *event.Context, pk packet.Packet) error {
	handler, ok := s.handlers[pk.ID()]
	if !ok {
		////fmt.printf("unhandled packet %T%v from %v\n", pk, fmt.Sprintf("%+v\n", pk)[1:], s.conn.RemoteAddr())
		return nil
	}
	if handler == nil {
		// A nil handler means it was explicitly unhandled.
		return nil
	}
	if err := handler.Handle(ctx, pk, s); err != nil {
		return fmt.Errorf("%T: %w", pk, err)
	}
	return nil
}

func (s *Session) sendMovement() {
	t := time.NewTicker(time.Millisecond * 25)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			if s.remoteConn == nil {
				continue
			}
			_ = s.remoteConn.WritePacket(&packet.PlayerAuthInput{
				Position: s.c.Position(),
				Pitch:    s.c.Rotation()[0],
				Yaw:      s.c.Rotation()[1],
			})
		}
	}
}

// handlePackets continuously handles incoming packets from the connection. It processes them accordingly.
// Once the connection is closed, handlePackets will return.
func (s *Session) handlePackets() {
	defer func() {
		// If this function ends up panicking, we don't want to call s.Close() as it may cause the entire
		// server to freeze without printing the actual panic message.
		// Instead, we check if there is a panic to recover, and just propagate the panic if this does happen
		// to be the case.
		if err := recover(); err != nil {
			panic(err)
		}
	}()
	for {
		pk, err := s.conn.ReadPacket()
		if err != nil {
			return
		}
		ctx := event.C()
		if err := s.handlePacket(ctx, pk); err != nil {
			// An error occurred during the handling of a packet. Print the error and stop handling any more
			// packets.
			//fmt.printf("failed processing packet from %v (%v): %v\n", s.conn.RemoteAddr(), s.c.Name(), err)
		}
		if s.remoteConn != nil {
			if !ctx.Cancelled() {
				if err = s.remoteConn.WritePacket(pk); err != nil {
					fmt.Printf("failed writing packet to %v (%v): %v\n", s.remoteConn.RemoteAddr(), s.c.Name(), err)
					return
				}
			}
			if pk, ok := pk.(*packet.PlayerAuthInput); ok {
				s.c.SetPosition(pk.Position)
				s.c.Rotate(mgl32.Vec2{pk.Pitch, pk.Yaw})
			}
		}
	}
}

// registerHandlers registers all packet handlers found in the packetHandler package.
func (s *Session) registerHandlers() {
	s.handlers = map[uint32]packetHandler{
		packet.IDModalFormResponse: &ModalFormResponseHandler{forms: make(map[uint32]form.Form)},
		packet.IDText:              &TextHandler{},
		packet.IDEmote:             &EmoteHandler{},
	}
}

// Close closes the session, which in turn closes the controllable and the connection that the session
// manages. Close ensures the method only runs code on the first call.
func (s *Session) Close() error {
	err := s.conn.Close()
	return err
}
