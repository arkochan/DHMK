package main

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	json "github.com/json-iterator/go"
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

func convertMessage(msg []byte, message *Message) error {
	// Unmarshal the JSON into msgIntermediate
	if err := json.Unmarshal(msg, message); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Check if required fields are present
	if message.Category == "" {
		return fmt.Errorf("category is required")
	}
	if message.Action == "" {
		return fmt.Errorf("action is required")
	}

	// for cases tht require body
	switch message.Action {
	case ActionTrade:
		if message.Body == nil {
			return fmt.Errorf("body is required for trade action")
		}
		bodyBytes, err := json.Marshal(message.Body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %v", err)
		}
		bodyStr := string(bodyBytes)
		// Use bodyStr as needed

		var gameTradeBody GameTradeBody
		if err := json.Unmarshal([]byte(bodyStr), &gameTradeBody); err != nil {
			return fmt.Errorf("failed to unmarshal body into GameTradeBody: %w", err)
		}

		// Replace the Body with the parsed GameTradeBody
		message.Body = gameTradeBody
	}
	// Assign the processed message to the output parameter
	return nil
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
		err = convertMessage(msg, &message)
		if err != nil {
			cr.MessagePlayer(name, "invalid message format")
			fmt.Println("Error:", err)
		}

		switch message.Category {
		// TODO:
		// Implement message type handling
		case CategoryGame:
			// Handle game messages
			// if cr.Board.CurrentPlayer().Name == name {
			broadcastMessage, promptMessage, err := cr.Board.HandleAction(player, message)
			if err != nil {
				cr.MessagePlayer(name, err.Error())
			}
			if broadcastMessage != "" {
				cr.MessageAll(broadcastMessage)
			}
			if promptMessage != "" {
				cr.MessagePlayer(name, promptMessage)
			}
		case CategoryRoom:
		// Handle room messages
		default:
			// Handle default messages
			fmt.Println(fmt.Errorf("Category %s Not defined", message.Category))
		}
	}
}
