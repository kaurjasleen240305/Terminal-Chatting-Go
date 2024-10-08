package websocket

import (
	"encoding/json"

	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/database"
	"gorm.io/gorm"
)

type Hub struct {
	Clients map[*Client]bool

	broadcast chan []byte

	register chan *Client

	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub {
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run(db *gorm.DB) {
	for {
		select {
		 case client := <-h.register :
			h.Clients[client] = true

         case client := <-h.unregister:
			if _,ok := h.Clients[client]; ok {
				delete(h.Clients,client)
				close(client.send)
			}

		case message := <- h.broadcast:
            var msg database.Message
			_ = json.Unmarshal(message,&msg)
			db.Create(&msg)
			for client := range h.Clients {
				if (msg.To == "" ||
					msg.To == client.Username ||
					msg.Username == client.Username) &&
					client.RoomCode == msg.RoomCode {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(h.Clients, client)
					}
				}
			}
		}
	}
}