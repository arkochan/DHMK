package di

import (
	"dhmk/delivery/handler/api"
	"dhmk/delivery/router"
	"dhmk/domain/repository"
	"dhmk/domain/service"
)

func GetRoomHandler(r *router.Router) *api.RoomHandler {
	room_repo := repository.NewRoomRepo()
	room_service := service.NewRoomService(room_repo)
	room_handler := api.NewRoomHandler(room_service)
	r.SetUpRoomRoutes(room_handler)
	return room_handler
}
