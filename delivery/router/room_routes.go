package router

import "dhmk/delivery/handler/api"

func (r *Router) SetUpRoomRoutes(room_handler *api.RoomHandler) {
	r.Engine.GET("/create", room_handler.CreateRoom())
}
