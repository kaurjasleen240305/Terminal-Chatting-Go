package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/database"
	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/routes"
	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/websocket"
	"github.com/gorilla/mux"
)

var addr = flag.String("addr", "localhost:8080", `http service address
	EXAMPLE:  ./server -addr localhost:5000
`)

func main() {
	flag.Parse()
	log.SetFlags((log.Lshortfile))
	hub := websocket.NewHub()
	db := database.NewDB()
	go hub.Run(db)
	r := mux.NewRouter()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})
	r.HandleFunc("/chat/{room}", func(w http.ResponseWriter, r *http.Request) {
		routes.ChatHandler(w, r, db)
	})
	r.HandleFunc("/online-users/{room}", func(w http.ResponseWriter, r *http.Request) {
		routes.OnlineUserHandler(w, r, hub)
	})
	r.HandleFunc("/valid-username/{room}", func(w http.ResponseWriter, r *http.Request) {
		routes.ValidUsernameHandler(w, r, hub)
	})
	http.Handle("/",r)
	log.Println(fmt.Sprintf("Server started at %s", *addr))
	log.Fatal(http.ListenAndServe(*addr, nil))
}