package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/database"
)

type OnlineUser struct {
	Username string `json:"username"`
	Color    int    `json:"color"`
}

var (
	Users chan OnlineUser
)

func GetInitialUsers(serverAddr string, roomCode string) {
	u := url.URL{Scheme: "http", Host: serverAddr, Path: fmt.Sprintf("/online-users/%s", roomCode)}
	res, err := http.Get(u.String())
	if err != nil {
		log.Println("error: GetInitialUsers couldn't fetch data")
		panic(0)
	}
	var data []OnlineUser
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println("error: GetInitialUsers couldn't parse data")
		panic(0)
	}
	Users = make(chan OnlineUser, 100)
	for _, user := range data {
		if user.Username != "" {
			Users <- user
		}
	}
}

func IsValidUsername(username string, serverAddr string, roomCode string) error {
	if username == "" {
		return errors.New("username must not be empty")
	}
	arr := strings.Split(username, " ")
	if len(arr) != 1 {
		return errors.New("username must not contain any space")
	}
	if serverAddr == "" {
		return nil
	}
	u := url.URL{Scheme: "http", Host: serverAddr, Path: fmt.Sprintf("/valid-username/%s", roomCode)}
	b, _ := json.Marshal(map[string]string{"username": username})
	res, err := http.Post(u.String(), "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println("error: isValidUsername couldn't fetch data")
		panic(0)
	}
	var data map[string]bool
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		panic(0)
	}
	if data["valid"] {
		return nil
	} else {
		return errors.New("username already taken!")
	}
}

func GetChat(serverAddr string, roomCode string) []database.Message {
	u := url.URL{Scheme: "http", Host: serverAddr, Path: fmt.Sprintf("/chat/%s", roomCode)}
	res, err := http.Get(u.String())
	if err != nil {
		log.Println("error: isValidUsername couldn't fetch data")
		panic(0)
	}
	var data []database.Message
	body, err := io.ReadAll(res.Body)
	err = json.Unmarshal(body, &data)
	if err != nil {
		log.Println(err)
		panic(0)
	}
	return data
}

func AddUser(newUser OnlineUser) {
	Users <- newUser
}

func RemoveUser(staleUser OnlineUser) {
	staleUser.Username = " " + staleUser.Username
	Users <- staleUser
}
