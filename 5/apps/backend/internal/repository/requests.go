package repository

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
)

var retryStrategy = retry.Strategy{
	Attempts: 3,
	Delay:    time.Second,
	Backoff:  math.Exp(1),
}

func CreateEvent(db *dbpg.DB, event *Event) (*Event, error) {
	tx, err := db.Master.Begin()
	if err != nil {
		return nil, err
	}

	query := `
INSERT INTO events VALUES ($1, $2, $3, $4);
`
	ctx := context.Background()

	_, err = tx.ExecContext(ctx, query, event.ID, event.Title, event.Date, event.AmountOfSeats)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	zlog.Logger.Info().Msg("Event inserted in db")

	queryForSeats := `
INSERT INTO seats VALUES ($1, $2, $3, $4);
`

	for i := 0; i < event.AmountOfSeats; i++ {
		seat := Seat{
			ID:       uuid.New(),
			IsBooked: false,
			IsPaid:   false,
		}
		event.Seats = append(event.Seats, seat)

		_, err = tx.ExecContext(ctx, queryForSeats, seat.ID, seat.IsBooked, seat.IsPaid, event.ID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	zlog.Logger.Info().Msg("Seats inserted in db")

	return event, nil
}

func ChangeBookSeat(db *dbpg.DB, seat *Seat) error {
	query := `
UPDATE seats SET is_booked = $1 WHERE uuid = $2;
`

	ctx := context.Background()

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, seat.IsBooked, seat.ID)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Seat updated in db")

	return nil
}

func PayBook(db *dbpg.DB, seat *Seat) error {
	query := `
UPDATE seats SET is_paid = TRUE WHERE uuid = $1;
`

	ctx := context.Background()

	_, err := db.ExecWithRetry(ctx, retryStrategy, query, seat.ID)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Seat updated in db")

	return nil
}

func GetEvent(db *dbpg.DB, id string) (*Event, error) {
	query := `
SELECT * FROM events WHERE uuid = $1;
`

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, id)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = row.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	var event Event

	for row.Next() {
		err = row.Scan(
			&event.ID,
			&event.Title,
			&event.Date,
			&event.AmountOfSeats,
		)
		if err != nil {
			return nil, err
		}
	}

	queryForSeats := `
SELECT * FROM seats WHERE event_id = $1
ORDER BY uuid;
`

	rowForSeats, err := db.QueryWithRetry(ctx, retryStrategy, queryForSeats, id)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = rowForSeats.Close()
		if err != nil {
			zlog.Logger.Error().Err(err)
		}
	}()

	for rowForSeats.Next() {
		seat := Seat{}
		eventId := ""

		err = rowForSeats.Scan(
			&seat.ID,
			&seat.IsBooked,
			&seat.IsPaid,
			&eventId,
		)
		if err != nil {
			return nil, err
		}
		event.Seats = append(event.Seats, seat)
	}

	zlog.Logger.Info().Msgf("Got event and %d seats", event.AmountOfSeats)

	return &event, nil
}

func GetSeatByID(db *dbpg.DB, id string) (*Seat, error) {
	query := `
SELECT * FROM seats WHERE uuid = $1;
`

	ctx := context.Background()

	row, err := db.QueryWithRetry(ctx, retryStrategy, query, id)
	if err != nil {
		return nil, err
	}

	seat := &Seat{}

	for row.Next() {
		eventId := ""

		err = row.Scan(
			&seat.ID,
			&seat.IsBooked,
			&seat.IsPaid,
			&eventId,
		)
		if err != nil {
			return nil, err
		}
	}

	return seat, nil
}
