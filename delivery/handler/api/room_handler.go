package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
}

func NewRoomHandler() *RoomHandler {
	return &RoomHandler{}
}

func (h *RoomHandler) CreateRoom() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"roomKey": "randomRoomKey"})
	}
}
