package server

import (
	"encoding/json"
	"fmt"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/mymaqc/proxy/form"
	"github.com/mymaqc/proxy/session"
	"github.com/mymaqc/proxy/user"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/sandertv/gophertunnel/minecraft/resource"
	"github.com/sandertv/gophertunnel/minecraft/text"
	"golang.org/x/oauth2"
	"strconv"
	"strings"
	"sync"
	"time"
)

type redirectionData struct {
	address     string
	tokenSource oauth2.TokenSource
}

type Server struct {
	addrMu sync.Mutex
	addr   map[string]redirectionData

	listener *minecraft.Listener
}

type statusProvider struct{}

func (statusProvider) ServerStatus(playerCount int, maxPlayers int) minecraft.ServerStatus {
	return minecraft.ServerStatus{
		ServerName:  "Proxy",
		PlayerCount: 68,
		MaxPlayers:  69,
	}
}

func New() *Server {
	s := &Server{
		addr:     make(map[string]redirectionData),
	}
	s.Listen(nil)
	return s
}

func (srv *Server)Listen(packs []*resource.Pack) {
	if srv.listener != nil {
		_ = srv.listener.Close()
	}
	l, err := minecraft.ListenConfig{
		StatusProvider: statusProvider{},
		ResourcePacks: packs,
		}.Listen("raknet", ":19132")
	if err != nil {
		panic(err)
	}
	srv.listener = l
}

func (srv *Server) Start() {
	fmt.Println("Proxy Started")
	for {
		c, err := srv.listener.Accept()
		if err != nil {
			continue
		}
		go srv.handleConn(c.(*minecraft.Conn), srv.listener)
	}
}

func (srv *Server)Packs(addr string, tkn oauth2.TokenSource) []*resource.Pack{
	c, err := minecraft.Dialer{
		TokenSource: tkn,
		}.Dial("raknet", addr)
	if err != nil {
		fmt.Println("couldn't download texture pack from", addr)
		return nil
	}
	c.Close()
	return c.ResourcePacks()
}

func (srv *Server) Redirect(conn *minecraft.Conn, addr string, tkn oauth2.TokenSource) {
	srv.addrMu.Lock()
	srv.addr[conn.IdentityData().DisplayName] = redirectionData{
		address:     addr,
		tokenSource: tkn,
	}
	srv.addrMu.Unlock()
	go func() {
		if addr == "localhost:19132" {
			return
		}
		ch := make(chan struct{})
		go func() {
			periodCount := 0
			t := time.NewTicker(time.Second)
			for {
				select {
				case <-t.C:
					if periodCount == 3 {
						periodCount = 0
					}
					periodCount++
					fmt.Println(periodCount)
					conn.WritePacket(&packet.SetTitle{
						ActionType: packet.TitleActionSetTitle,
						Text: fmt.Sprintf("§adownloading pack for \n §6%s§a%s",addr, strings.Repeat(".", periodCount)),
						})
					case <-ch:
						t.Stop()
						return
				}
			}
		}()

		packs := srv.Packs(addr, tkn)
		ch <- struct{}{}
		_ = conn.WritePacket(&packet.Transfer{
			Address: "localhost",
			Port:    19132,
			})
		if packs != nil {
			time.Sleep(time.Second / 4)
			srv.Listen(packs)
		}
	}()
	time.AfterFunc(time.Second*60, func() {
		srv.addrMu.Lock()
		delete(srv.addr, conn.IdentityData().DisplayName)
		srv.addrMu.Unlock()
	})
}
func (srv *Server) handleConn(conn *minecraft.Conn, listener *minecraft.Listener) {
	srv.addrMu.Lock()
	addr, ok := srv.addr[conn.IdentityData().DisplayName]
	srv.addrMu.Unlock()
	if !ok {
		_ = conn.StartGame(minecraft.GameData{})
		s := session.New(conn, minecraft.GameData{})
		u := user.New(s)

		s.SetControllable(u)
		u.SendForm(form.NewTransfer())

		addr, tkn := u.WaitForTransfer()
		srv.Redirect(conn, addr, tkn)
		return
	}
	clientData := conn.ClientData()
	clientData.DeviceModel = strings.ToUpper(clientData.DeviceModel)
	serverConn, err := minecraft.Dialer{
		TokenSource: addr.tokenSource,
		ClientData:  clientData,
	}.Dial("raknet", addr.address)
	if err != nil {
		_ = conn.StartGame(minecraft.GameData{})
		_ = conn.WritePacket(&packet.Text{
			TextType: packet.TextTypeChat,
			Message:  text.Colourf("<red>%v</red>", err),
		})

		time.Sleep(time.Second * 5)
		_ = conn.WritePacket(&packet.Transfer{
			Address: "localhost",
			Port:    19132,
		})
		srv.addrMu.Lock()
		delete(srv.addr, conn.IdentityData().DisplayName)
		srv.addrMu.Unlock()
		return
	}

	d := serverConn.GameData()
	s := session.New(conn, d)
	u := user.New(s)
	s.SetControllable(u)

	var g sync.WaitGroup
	g.Add(2)
	go func() {
		_ = conn.StartGame(d)
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			fmt.Println(err)
			return
		}
		g.Done()
	}()
	g.Wait()
	s.SetRemoteConn(serverConn)
	go func() {
		defer serverConn.Close()
		for {
			pk, err := serverConn.ReadPacket()
			if err != nil {
				return
			}
			switch pk := pk.(type) {
			case *packet.ModalFormRequest:
				if u.Settings().SliderExploit {
					f := &frm{}
					err := json.Unmarshal(pk.FormData, f)
					if err == nil && f.Type == "custom_form" {
						for i, e := range f.Content {
							for _, t := range e {
								if t == "slider"{
									f.Content[i]["max"] = 1000
									f.Content[i]["min"] = -1000
									fmt.Println(f.Content[i])
								}else if t == "step_slider" {
									n := make([]string, 200)
									for j := range n {
										nb := j+1
										if j < 100 {
											nb = -nb
										}
										n[j] = strconv.Itoa(nb)
									}
									f.Content[i]["default"] = 0
									f.Content[i]["steps"] = n
									fmt.Println(f)
								}
							}
						}
						d, _ := f.MarshalJSON()
						_ = conn.WritePacket(&packet.ModalFormRequest{
							FormID:   pk.FormID,
							FormData: d,
						})
						continue
					}
				}
			case *packet.MovePlayer:
				if pk.EntityRuntimeID == u.RuntimeID() {
					u.SetPosition(pk.Position)
					u.Rotate(mgl32.Vec2{pk.Pitch, pk.Yaw})
				}
			case *packet.MoveActorDelta:
				if pk.EntityRuntimeID == u.RuntimeID() {
					u.SetPosition(pk.Position)
				}
			case *packet.MoveActorAbsolute:
				if pk.EntityRuntimeID == u.RuntimeID() {
					u.SetPosition(pk.Position)
				}
			case *packet.Transfer:
				srv.Redirect(conn, fmt.Sprintf("%s:%d", pk.Address, pk.Port), addr.tokenSource)
				return
			}
			if err := conn.WritePacket(pk); err != nil {
				return
			}
		}
	}()
}

type frm struct {
	Title   string           `json:"title"`
	Type    string           `json:"type"`
	Content []map[string]any `json:"content"`
}

func (f frm) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"title":   f.Title,
		"type":    f.Type,
		"content": f.Content,
	}
	return json.Marshal(m)
}
