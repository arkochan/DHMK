package repository

import (
	"dhmk/domain/model"
	"fmt"

	"github.com/google/uuid"
)

type roomRepo struct {
	rooms map[string]model.Room // Maps room keys to Room objects
}

type RoomRepo interface {
	CreateRoom() *model.Room
	GetRoom(roomKey string) (*model.Room, error)
	DeleteRoom(roomKey string) error
	ListRooms() []*model.Room
	AddPlayerToRoom(roomKey string, player *model.Player) error
}

func NewRoomRepo() RoomRepo {
	return &roomRepo{
		rooms: make(map[string]model.Room),
	}
}

func (r *roomRepo) CreateRoom() *model.Room {
	var rkey string
	for {
		rkey = uuid.New().String()
		if _, exists := r.rooms[rkey]; !exists {
			break
		}
	}
	room := model.Room{
		RoomKey: rkey,
	}
	r.rooms[rkey] = room
	return &room
}

func (r *roomRepo) GetRoom(roomKey string) (*model.Room, error) {
	room, exists := r.rooms[roomKey]
	if !exists {
		return nil, fmt.Errorf("room with key %s not found", roomKey)
	}
	return &room, nil
}

func (r *roomRepo) DeleteRoom(roomKey string) error {
	if _, exists := r.rooms[roomKey]; !exists {
		return fmt.Errorf("room with key %s not found", roomKey)
	}
	delete(r.rooms, roomKey)
	return nil
}

func (r *roomRepo) ListRooms() []*model.Room {
	rooms := []*model.Room{}
	for _, room := range r.rooms {
		rooms = append(rooms, &room)
	}
	return rooms
}

func (r *roomRepo) AddPlayerToRoom(roomKey string, player *model.Player) error {
	room, exists := r.rooms[roomKey]
	if !exists {
		return fmt.Errorf("room with key %s not found", roomKey)
	}
	room.Players = append(room.Players, *player)
	r.rooms[roomKey] = room
	return nil
}
