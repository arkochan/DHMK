package main

import (
	"dhmk/delivery/router"
	"dhmk/domain/di"
)

func main() {
	router := router.NewRouter()
	di.DI(router)
	// r.GET("/ws/:roomKey", func(c *gin.Context) {
	// 	roomKey := c.Param("roomKey")
	// 	roomInstance := room.GetOrCreateRoom(roomKey)
	// 	roomInstance.HandleWebSocket(c)
	// })

	// r.LoadHTMLGlob("templates/*")
	// r.GET("/room/:roomKey", func(c *gin.Context) {
	// 	roomKey := c.Param("roomKey")
	// 	c.HTML(http.StatusOK, "index.html", gin.H{"roomKey": roomKey})
	// })

	// r.GET("/", func(c *gin.Context) {
	// 	c.HTML(http.StatusOK, "index.html", nil)
	// })

	// r.GET("/newroom", func(c *gin.Context) {
	// 	roomKey := room.CreateRandomRoom()
	// 	c.JSON(http.StatusOK, gin.H{"roomKey": roomKey})
	// })

	// r.Run(":8080")
}
