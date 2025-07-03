package api

import (
	"dhmk/domain/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	room_service *service.RoomService
}

func NewRoomHandler(s *service.RoomService) *RoomHandler {
	return &RoomHandler{
		room_service: s,
	}
}

func (h *RoomHandler) CreateRoomHandler() gin.HandlerFunc {
	newRoomKey := h.room_service.CreateRoom()
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"roomKey": newRoomKey})
	}
}
