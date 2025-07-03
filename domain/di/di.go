package di

import (
	"dhmk/delivery/router"
	"fmt"
)

func DI(r *router.Router) {
	// Initialize the room handler and set up routes
	GetRoomHandler(r)

	// You can add more handlers and their routes here as needed
	// For example:
	// GetUserHandler(r)
	// GetMessageHandler(r)
	r.Engine.Run(":8090") // Start the server on port 8080
	fmt.Println("Server running on port 8090")

}
