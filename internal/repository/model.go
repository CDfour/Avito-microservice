package repository

import (
	"time"

	"github.com/google/uuid"
)

type user struct {
	id         uuid.UUID
	balance    float64
	dateCreate time.Time
	lastUpdate time.Time
}

type order struct {
	id          uuid.UUID
	userID      uuid.UUID
	serviceID   uuid.UUID
	serviceName string
	dateCreate  time.Time
	funds       float64
}

type history struct {
	id          uuid.UUID
	serviceName string
	cost        float64
	date        time.Time
}
