package routes

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/database"
	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/user"
	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/websocket"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func OnlineUserHandler(w http.ResponseWriter, r *http.Request, hub *websocket.Hub) {
	if r.Method != "GET" {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	onlineUsers := []user.OnlineUser{}
	for client := range hub.Clients {
		if client.RoomCode == vars["room"] {
			onlineUsers = append(
				onlineUsers,
				user.OnlineUser{
					Username: (*client).Username,
					Color:    (*client).Color,
				},
			)
		}
	}
	jsonRes, err := json.Marshal(onlineUsers)
	if err != nil {
		log.Println(err)
	}
	w.Write(jsonRes)
}
func ValidUsernameHandler(w http.ResponseWriter, r *http.Request, hub *websocket.Hub) {
	if r.Method != "POST" {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot parse request body", http.StatusBadRequest)
	}
	var data map[string]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Cannot parse request body", http.StatusBadRequest)
	}
	for client := range hub.Clients {
		if client.RoomCode == vars["room"] && client.Username == data["username"] {
			res := map[string]bool{"valid": false}
			jsonRes, _ := json.Marshal(res)
			w.Write([]byte(jsonRes))
			return
		}
	}
	res := map[string]bool{"valid": true}
	jsonRes, _ := json.Marshal(res)
	w.Write([]byte(jsonRes))
}

func ChatHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	if r.Method != "GET" {
		http.Error(w, "Method not supported", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)

	var chat []database.Message
	db.Where(&database.Message{RoomCode: vars["room"]}).Find(&chat)

	jsonRes, _ := json.Marshal(chat)
	w.Write([]byte(jsonRes))
}
