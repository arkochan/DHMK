package dhmk

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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

		command := string(msg)
		if command == "go" && cr.Board.CurrentPlayer().Name == name {
			steps := rand.Intn(6) + 1
			cr.Board.MovePlayer(player, steps)
			cr.Board.NextTurn()
			update := fmt.Sprintf("%s rolled %d and moved to %s", name, steps, cr.Board.Slots[player.Position].Name)
			cr.Broadcast <- update
		} else {
			cr.Broadcast <- fmt.Sprintf("It's not %s's turn!", name)
		}
	}
}
