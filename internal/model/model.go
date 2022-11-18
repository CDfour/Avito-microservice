package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	Funds      float64
	DateCreate time.Time
	LastUpdate time.Time
}

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	ServiceID   uuid.UUID
	ServiceName string
	DateCreate  time.Time
	Funds       float64
}

type Report struct {
	ServiceName string
	Revenue     float64
}

type History struct {
	UserID      uuid.UUID
	ServiceName string
	Cost        float64
	OrderDate   time.Time
}
