package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Kost0/L3/internal/queue"
	"github.com/Kost0/L3/internal/repository"
	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
)

type Handler struct {
	DB       *dbpg.DB
	SeatChan chan repository.Seat
	Booking  *queue.Booking
}

func (h *Handler) CreateEvent(c *ginext.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	newEvent := &repository.EventDTO{}

	err = json.Unmarshal(data, newEvent)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	event := &repository.Event{
		ID:            uuid.New(),
		Title:         newEvent.Title,
		Date:          newEvent.Date,
		AmountOfSeats: newEvent.AmountOfSeats,
		Seats:         []repository.Seat{},
	}

	event, err = repository.CreateEvent(h.DB, event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"data": event})
}

func (h *Handler) Book(c *ginext.Context) {
	id := c.Param("id")

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	var number repository.SeatId

	err = json.Unmarshal(data, &number)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	event, err := repository.GetEvent(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if number.SeatIndex > event.AmountOfSeats {
		c.JSON(http.StatusBadRequest, ginext.H{"data": "seat number out of range"})
		return
	}

	seat := event.Seats[number.SeatIndex-1]

	if seat.IsBooked || seat.IsPaid {
		c.JSON(http.StatusConflict, ginext.H{"data": "seat already booked"})
		return
	}

	bookingRequest := repository.Seat{
		ID:          seat.ID,
		IsBooked:    true,
		IsPaid:      seat.IsPaid,
		BookedTime:  time.Now(),
		CancelTimer: make(chan struct{}),
	}

	err = repository.ChangeBookSeat(h.DB, &bookingRequest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	h.SeatChan <- bookingRequest

	c.JSON(http.StatusOK, ginext.H{"data": bookingRequest})
}

func (h *Handler) MakePayment(c *ginext.Context) {
	id := c.Param("id")

	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	var number repository.SeatId

	err = json.Unmarshal(data, &number)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	event, err := repository.GetEvent(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	if number.SeatIndex > event.AmountOfSeats {
		c.JSON(http.StatusBadRequest, ginext.H{"data": "seat number out of range"})
		return
	}

	seat := event.Seats[number.SeatIndex-1]

	err = repository.PayBook(h.DB, &seat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	h.Booking.CancelTimer(seat.ID.String())

	c.JSON(http.StatusOK, ginext.H{"data": seat})
}

func (h *Handler) GetEvent(c *ginext.Context) {
	id := c.Param("id")

	event, err := repository.GetEvent(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	eventResponse := &repository.EventResponse{
		ID:             event.ID,
		Title:          event.Title,
		Date:           event.Date,
		AmountOfSeats:  event.AmountOfSeats,
		AvailableSeats: 0,
		Seats:          make([]repository.SeatDTO, 0),
	}

	for i, seat := range event.Seats {
		status := ""
		if seat.IsPaid {
			status = "paid"
		} else if seat.IsBooked {
			status = "reserved"
		} else {
			eventResponse.AvailableSeats++
			status = "free"
		}
		seatDTO := repository.SeatDTO{
			Index:  i + 1,
			Status: status,
		}
		eventResponse.Seats = append(eventResponse.Seats, seatDTO)
	}

	c.JSON(http.StatusOK, eventResponse)
}
