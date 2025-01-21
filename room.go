package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type (
	Category string
	Action   string
)

const (
	CategoryGame Category = "game"
	CategoryRoom Category = "room"

	ActionGo          Action = "go"
	ActionTrade       Action = "trade"
	ActionMessage     Action = "message"
	ActionUseCard     Action = "useCard"
	ActionForfeitGame Action = "forfeit"
	ActionMortgage    Action = "mortgage"
	ActionBuyHouse    Action = "house"
	ActionEndTurn     Action = "end"
	ActionBuy         Action = "buy"
)

type Message struct {
	Category Category    `json:"category"`
	Action   Action      `json:"action"`
	Body     interface{} `json:"body,omitempty"`
}

type RoomMessageBody struct {
	Message string `json:"body"`
}

type Room struct {
	Board     *Board
	Clients   map[*websocket.Conn]string
	Broadcast chan string
	sync.Mutex
}

func NewRoom() *Room {
	return &Room{
		Board:     NewBoard(),
		Clients:   make(map[*websocket.Conn]string),
		Broadcast: make(chan string),
	}
}

func (cr *Room) Run() {
	for {
		msg := <-cr.Broadcast
		cr.Lock()
		for client := range cr.Clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				client.Close()
				delete(cr.Clients, client)
			}
		}
		cr.Unlock()
	}
}

func (cr *Room) MessageAll(message string) {
	cr.Broadcast <- message
}

func (cr *Room) MessagePlayer(player string, message string) {
	for conn, name := range cr.Clients {
		if name == player {
			conn.WriteMessage(websocket.TextMessage, []byte(message))
			break
		}
	}
}

func (cr *Room) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	defer func() {
		cr.Lock()
		delete(cr.Clients, conn)
		cr.Unlock()
		conn.Close()
	}()

	// Get player name from query
	name := c.Query("name")
	if name == "" {
		// Assign player name as Player-N instead
		name = fmt.Sprintf("Player-%d", len(cr.Clients)+1)
	}

	player := cr.Board.AddPlayer(name)
	cr.Lock()
	cr.Clients[conn] = name
	cr.Unlock()

	cr.Broadcast <- fmt.Sprintf("%s joined the game!", name)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Read error:", err)
			break
		}

		var message Message
		json.Unmarshal(msg, &message)
		}
	}
}
