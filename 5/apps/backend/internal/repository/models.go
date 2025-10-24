package repository

import (
	"time"

	"github.com/google/uuid"
)

type Event struct {
	ID            uuid.UUID `json:"uuid"`
	Title         string    `json:"title"`
	Date          time.Time `json:"date"`
	AmountOfSeats int       `json:"amount_of_seats"`
	Seats         []Seat    `json:"-"`
}

type Seat struct {
	ID          uuid.UUID     `json:"uuid"`
	IsBooked    bool          `json:"is_booked"`
	IsPaid      bool          `json:"is_paid"`
	BookedTime  time.Time     `json:"-"`
	CancelTimer chan struct{} `json:"-"`
}

type EventDTO struct {
	Title         string    `json:"title"`
	Date          time.Time `json:"date"`
	AmountOfSeats int       `json:"amount_of_seats"`
}

type SeatDTO struct {
	Index  int    `json:"index"`
	Status string `json:"status"`
}

type EventResponse struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Date           time.Time `json:"date"`
	AmountOfSeats  int       `json:"amount_of_seats"`
	AvailableSeats int       `json:"available_seats"`
	Seats          []SeatDTO `json:"seats"`
}

type SeatId struct {
	SeatIndex int `json:"seat_index"`
}
