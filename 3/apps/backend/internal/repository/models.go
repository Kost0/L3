package repository

import (
	"github.com/google/uuid"
)

type Comment struct {
	UUID   *uuid.UUID `json:"id"`
	Text   string     `json:"text"`
	Parent *uuid.UUID `json:"parent"`
	Vector string     `json:"search_vector"`
}
