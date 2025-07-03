package model

type Room struct {
	RoomKey string   `json:"roomKey"`
	Players []Player `json:"players"`
}
