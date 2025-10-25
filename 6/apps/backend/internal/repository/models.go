package repository

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID  uuid.UUID `json:"uuid"`
	Title    string    `json:"title"`
	Cost     int       `json:"cost"`
	Items    int       `json:"items"`
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
}

type OrderDTO struct {
	Title    string    `json:"title"`
	Cost     int       `json:"cost"`
	Items    int       `json:"items"`
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
}
