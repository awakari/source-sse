package model

import "time"

type Stream struct {
	Auth      string
	GroupId   string
	UserId    string
	CreatedAt time.Time
	Replica   uint32
}
