package service

import "dhmk/domain/repository"

type RoomService struct {
	RoomRepo repository.RoomRepo
}

func NewRoomService(roomRepo repository.RoomRepo) *RoomService {
	return &RoomService{
		RoomRepo: roomRepo,
	}
}

func (s *RoomService) CreateRoom() string {
	// Logic to create a room and return the room key
	// For now, we return a dummy room key
	room := s.RoomRepo.CreateRoom()
	return room.RoomKey
}
