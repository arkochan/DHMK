// Package room manages game rooms, WebSocket connections, and message routing.
package room

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"dhmk/game"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	json "github.com/json-iterator/go"
)

// Category represents the type of message (game or room).
type (
	Category string
	Action   string
)

const (
	CategoryGame Category = "game"
	CategoryRoom Category = "room"

	ActionGo          Action = "go"
	ActionTrade       Action = "trade"
	ActionAcceptTrade Action = "acceptTrade"
	ActionMessage     Action = "message"
	ActionUseCard     Action = "useCard"
	ActionForfeitGame Action = "forfeit"
	ActionMortgage    Action = "mortgage"
	ActionBuyHouse    Action = "house"
	ActionEndTurn     Action = "end"
	ActionBuy         Action = "buy"
)

// Message represents a message sent between client and server over WebSocket.
type Message struct {
	Category Category    `json:"category"`
	Action   Action      `json:"action"`
	Body     interface{} `json:"body,omitempty"`
}

// RoomMessageBody is used for simple room messages.
type RoomMessageBody struct {
	Message string `json:"body"`
}

// Room manages the game state, connected clients, and message broadcasting.
type Room struct {
	Board     *game.Board
	Clients   map[*websocket.Conn]string
	Broadcast chan string
	sync.Mutex
}

var (
	rooms   = make(map[string]*Room)
	roomsMu sync.RWMutex

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// NewRoom creates and returns a new Room instance.
func NewRoom() *Room {
	return &Room{
		Board:     game.NewBoard(),
		Clients:   make(map[*websocket.Conn]string),
		Broadcast: make(chan string),
	}
}

// GetOrCreateRoom returns the Room for a key, creating it if necessary.
func GetOrCreateRoom(key string) *Room {
	roomsMu.Lock()
	defer roomsMu.Unlock()
	room, ok := rooms[key]
	if !ok {
		room = NewRoom()
		rooms[key] = room
		go room.Run()
	}
	return room
}

// Run listens for broadcast messages and sends them to all connected clients.
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

// MessageAll sends a message to all connected clients.
func (cr *Room) MessageAll(message string) {
	cr.Broadcast <- message
}

// MessagePlayer sends a message to a specific player by name.
func (cr *Room) MessagePlayer(player string, message string) {
	for conn, name := range cr.Clients {
		if name == player {
			conn.WriteMessage(websocket.TextMessage, []byte(message))
			break
		}
	}
}

// getBodyStr marshals the body to a JSON string for further processing.
func getBodyStr(body interface{}) (string, error) {
	if body == nil {
		return "", fmt.Errorf("body is required for trade action")
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal body: %v", err)
	}
	bodyStr := string(bodyBytes)
	// Use bodyStr as needed
	//
	return bodyStr, nil
}

// convertMessage parses and converts incoming WebSocket messages to the correct Go struct based on action.
// It validates required fields and, for certain actions (trade/acceptTrade),
// further unmarshals the Body into the appropriate struct (GameTradeBody or GameTradeAcceptBody).
// Returns an error if the message is invalid or cannot be parsed.
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
		bodyStr, err := getBodyStr(message.Body)
		if err != nil {
			return fmt.Errorf("failed to get body string: %w", err)
		}

		var gameTradeBody game.GameTradeBody
		if err := json.Unmarshal([]byte(bodyStr), &gameTradeBody); err != nil {
			return fmt.Errorf("failed to unmarshal body into GameTradeBody: %w", err)
		}

		// Replace the Body with the parsed GameTradeBody
		message.Body = gameTradeBody
	case ActionAcceptTrade:
		bodyStr, err := getBodyStr(message.Body)
		if err != nil {
			return fmt.Errorf("failed to get body string: %w", err)
		}

		var gameTradeAcceptBody game.GameTradeAcceptBody

		if err := json.Unmarshal([]byte(bodyStr), &gameTradeAcceptBody); err != nil {
			return fmt.Errorf("failed to unmarshal body into GameTradeAcceptBody: %w", err)
		}
		// Replace the Body with the parsed GameTradeBody
		message.Body = gameTradeAcceptBody
	}
	// Assign the processed message to the output parameter
	return nil
}

// HandleWebSocket upgrades the HTTP connection to a WebSocket, registers the player, and processes incoming messages.
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
		case CategoryGame:
			// Handle game messages
			// if cr.Board.CurrentPlayer().Name == name {
			// Map Action type to string for game.HandleAction
			actionString := ""
			var body interface{}
			switch message.Action {
			case ActionTrade:
				actionString = "trade"
				body = message.Body
			case ActionAcceptTrade:
				actionString = "accept_trade"
				body = message.Body
			case ActionForfeitGame:
				actionString = "forfeit_game"
				body = nil
			case ActionGo:
				actionString = "go"
				body = nil
			case ActionBuy:
				actionString = "buy"
				body = nil
			case ActionEndTurn:
				actionString = "end_turn"
				body = nil
			// Add more actions as needed
			default:
				cr.MessagePlayer(name, "invalid action")
				continue
			}

			broadcastMessage, promptMessage, err := cr.Board.HandleAction(player, actionString, body)
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

// CreateRandomRoom generates a random room key, creates the room, and returns the key.
func CreateRandomRoom() string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	keyLen := 6
	roomKeyBytes := make([]byte, keyLen)
	for i := range roomKeyBytes {
		roomKeyBytes[i] = letters[randInt(len(letters))]
	}
	roomKey := string(roomKeyBytes)
	GetOrCreateRoom(roomKey)
	return roomKey
}

// randInt returns a random int in [0, n)
func randInt(n int) int {
	return int(time.Now().UnixNano() % int64(n))
}
