package repository

import (
	"time"

	"github.com/google/uuid"
)

type Item struct {
	UUID      uuid.UUID `json:"uuid"`
	Title     string    `json:"title"`
	Price     float64   `json:"price"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ItemDTO struct {
	Title    string  `json:"title"`
	Price    float64 `json:"price"`
	Category string  `json:"category"`
}

type History struct {
	ItemID        uuid.UUID `json:"item_id"`
	OperationType string    `json:"operation_type"`
	RoleName      string    `json:"role_name"`
	ChangedAt     time.Time `json:"changed_at"`
	//Difference    string    `json:"difference"`
}
