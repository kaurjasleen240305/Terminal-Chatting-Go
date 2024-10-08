package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"
	"os/signal"
	"regexp"
	"encoding/json"
	"strings"

	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/color"
	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/user"
	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/database"
	"github.com/kaurjasleen240305/Terminal-Chatting-Go/internal/textParser"
	"github.com/gorilla/websocket"
)


func printMsg(msg database.Message) {
	nameContent := msg.Username
	if msg.To != "" {
		if msg.Username == *username {
			nameContent = nameContent + "(" + msg.To + ")"
		} else {
			nameContent = nameContent + "(private)"
		}
	}
	content := textParser.Parse(msg.Content)
	time := color.Grey(msg.Time)
	if msg.Content == "" {
		return
	} else if err := user.IsValidUsername(msg.Username, "", *roomCode); err != nil {
		fmt.Printf("%s %s\n", time, content)
	} else if msg.Username == *username {
		fmt.Printf("%s %s: %s\n", time, color.CustomWithBg(nameContent, msg.Color), content)
	} else {
		fmt.Printf("%s %s: %s\n", time, color.Custom(nameContent, msg.Color), content)
	}
}

func sendAnnouncement(
	username string,
	announcementType string,
	writeMessage func(messageType int, data []byte) error,
) error {
	var newMsg database.Message
	customUserName := " y"+username
	if announcementType == "left"{
		customUserName = " x"+username
	}
	t := time.Now()
	newMsg = database.Message{
		Username: customUserName,
		Content:  username+" "+announcementType+" the chat!",
		Color:    myColor,
		Time:     fmt.Sprintf("%d:%d:%d", t.Hour(), t.Minute(), t.Second()),
		To:       "",
		RoomCode: *roomCode,
	}
	postBody, _ := json.Marshal(newMsg)
	err := writeMessage(websocket.TextMessage, []byte(postBody))
	return err
}

func sendMsg(content string, writeMessage func(messageType int, data []byte) error) error {
	t := time.Now()
	newMsg := database.Message{
		Username: *username,
		Content:  content,
		Color:    myColor,
		Time:     fmt.Sprintf("%d:%d:%d", t.Hour(), t.Minute(), t.Second()),
		To:       "",
		RoomCode: *roomCode,
	}
	to, actualContent := inputParser(content)
	newMsg.To = to
	newMsg.Content = actualContent
	postBody, _ := json.Marshal(newMsg)
	err := writeMessage(websocket.TextMessage, []byte(postBody))
	return err
}

func sendMessageToServer(done <-chan struct{}, interrupt <-chan os.Signal, input <-chan string, c websocket.Conn) {
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			err := sendAnnouncement(*username, "left", c.WriteMessage)
			if err != nil {
				log.Println("err:", err)
				return
			}
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err = c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			os.Exit(0)
		case text := <-input:
			if text[len(text)-1] == '\n' {
				text = text[:len(text)-1]
			}
			err := sendMsg(text, c.WriteMessage)
			if err != nil {
				log.Println("write:", err)
				return
			}
		}
	}
}

func readSTDIN(input chan string, reader *bufio.Reader) {
	defer close(input)
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		input <- text
	}
}

func readMessageFromServer(done chan struct{}, c websocket.Conn) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("websocket error: ", err)
			return
		}
		var msg database.Message
		err = json.Unmarshal(message, &msg)
		if err != nil {
			log.Println("Parsing error:", err)
			return
		}
		if msg.Username[0] == ' ' {
			if msg.Username[1] == 'x' {
				user.RemoveUser(user.OnlineUser{
					Username: msg.Username[2:],
					Color:    msg.Color,
				})
			} else if msg.Username[1] == 'y' {
				user.AddUser(user.OnlineUser{
					Username: msg.Username[2:],
					Color:    msg.Color,
				})
			}
		}
		printMsg(msg)
	}
}

func inputParser(content string) (string, string) {
	if len(content) > 2 && content[0] == '>' {
		arr := strings.Split(content, " ")
		return arr[0][1:], strings.Join(arr[1:], " ")
	}
	return "", content
}

var (
	addr = flag.String("addr", "localhost:8080", `http service address
EXAMPLE: ./client -addr localhost:5000	
	`)
	username = flag.String("user", "Newbie", `username for chat
EXAMPLE: ./client -user aksh	
	`)
	roomCode = flag.String("room", "general", `Specify room code
EXAMPLE: ./client -room private	
	`)
	myColor int
)

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)
	if err := user.IsValidUsername(*username, *addr, *roomCode); err != nil {
		fmt.Println(err)
		return
	}
	roomCodeR, _ := regexp.Compile("[a-z]{3,6}")
	if !roomCodeR.MatchString(*roomCode) {
		fmt.Println("Room code must be 3 to 6 letter string with lowercase alphabets(a-z)")
		return
	}

	myColor = color.Random()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		panic(err)
	}
	defer c.Close()

	user.GetInitialUsers(*addr, *roomCode)
	done := make(chan struct{})
	input := make(chan string)

	for _, msg := range user.GetChat(*addr, *roomCode) {
			printMsg(msg)
	}

	// Welcome message
	err = sendAnnouncement(*username, "joined", c.WriteMessage)
	if err != nil {
		log.Println("err:", err)
		return
	}

	// Read messages
	go readMessageFromServer(done, *c)
	reader := bufio.NewReader(os.Stdin)
	go readSTDIN(input, reader)

	// Send message to server
	defer close(input)
	sendMessageToServer(done, interrupt, input, *c)
}
