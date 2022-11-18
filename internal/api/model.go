package api

import "github.com/google/uuid"

type message struct {
	Message string `json:"message"`
}

type user struct {
	ID    uuid.UUID `json:"id"`
	Funds float64   `json:"funds"`
}

type transfer struct {
	SenderID    uuid.UUID `json:"sender_id"`
	RecipientID uuid.UUID `json:"recipient_id"`
	Funds       float64   `json:"funds"`
}

type order struct {
	UserID      uuid.UUID `json:"user_id"`
	ServiceID   uuid.UUID `json:"service_id"`
	ServiceName string    `json:"service_name"`
	OrderID     uuid.UUID `json:"order_id"`
	Cost        float64   `json:"cost"`
}

type report struct {
	Year  string `json:"year"`
	Month string `json:"month"`
}
