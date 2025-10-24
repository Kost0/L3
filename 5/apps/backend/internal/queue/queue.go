package queue

import (
	"sync"
	"time"

	"github.com/Kost0/L3/internal/repository"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

const N = 10 * time.Second

type Booking struct {
	DB           *dbpg.DB
	activeTimers map[string]chan struct{}
	mu           sync.Mutex
}

func NewBooking(db *dbpg.DB) *Booking {
	return &Booking{
		DB:           db,
		activeTimers: make(map[string]chan struct{}),
	}
}

func (b *Booking) StartQueue(ch <-chan repository.Seat) {
	for seat := range ch {
		seatID := seat.ID.String()

		b.mu.Lock()
		b.activeTimers[seatID] = seat.CancelTimer
		b.mu.Unlock()

		go b.startTimerForSeat(seat)
	}
}

func (b *Booking) startTimerForSeat(seat repository.Seat) {
	seatID := seat.ID.String()

	defer func() {
		b.mu.Lock()
		delete(b.activeTimers, seatID)
		b.mu.Unlock()

		if seat.CancelTimer != nil {
			close(seat.CancelTimer)
		}
	}()

	zlog.Logger.Info().Msgf("Timer for seat %s started", seat.ID)
	timeForBooking := N - time.Since(seat.BookedTime)

	zlog.Logger.Info().Msgf("time to wait: %v", timeForBooking)

	if timeForBooking < 0 {
		timeForBooking = time.Second
	}

	timer := time.NewTimer(timeForBooking)
	defer timer.Stop()

	select {
	case <-timer.C:
		zlog.Logger.Info().Msg("Time is over")

		seatFromDB, err := repository.GetSeatByID(b.DB, seat.ID.String())
		if err != nil {
			zlog.Logger.Info().Msgf("Error getting seat from database: %v", err)
			return
		}

		seat.IsPaid = seatFromDB.IsPaid

		if !seat.IsPaid {
			seat.IsBooked = false
			err = repository.ChangeBookSeat(b.DB, &seat)
			if err != nil {
				zlog.Logger.Info().Msgf("Error updating seat: %v", err)
				return
			}
		}

		zlog.Logger.Info().Msg("Time ended")
	case <-seat.CancelTimer:
		zlog.Logger.Info().Msg("Book was paid")
		return
	}
}

func (b *Booking) CancelTimer(seatID string) bool {
	b.mu.Lock()
	cancelChan, exists := b.activeTimers[seatID]
	b.mu.Unlock()

	if !exists {
		zlog.Logger.Info().Msgf("No active timer for seat %s", seatID)
		return false
	}

	select {
	case cancelChan <- struct{}{}:
		zlog.Logger.Info().Msgf("Cancelled timer for seat %s", seatID)
		return true
	default:
		zlog.Logger.Info().Msgf("Timer channel full for seat %s", seatID)
		return false
	}
}
