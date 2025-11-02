package main

import (
	"github.com/Kost0/L3/internal/handlers"
	"github.com/Kost0/L3/internal/queue"
	"github.com/Kost0/L3/internal/repository"
	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

const (
	eventRoute  = "events"
	eventById   = "events/:id"
	bookSeat    = "events/:id/book"
	confirmBook = "events/:id/confirm"
)

func main() {
	zlog.Init()

	db, err := repository.ConnectDB()
	if err != nil {
		panic(err)
	}

	err = repository.RunMigrations(db, "notifications")
	if err != nil {
		panic(err)
	}

	zlog.Logger.Info().Msg("DB started")

	SeatChan := make(chan repository.Seat)

	booking := queue.NewBooking(db)

	go booking.StartQueue(SeatChan)

	handler := handlers.Handler{
		DB:       db,
		SeatChan: SeatChan,
		Booking:  booking,
	}

	engine := ginext.New("")

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5000"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Accept", "Content-Type"},
	}))

	engine.POST(eventRoute, handler.CreateEvent)

	engine.POST(bookSeat, handler.Book)

	engine.POST(confirmBook, handler.MakePayment)

	engine.GET(eventById, handler.GetEvent)

	err = engine.Run(":8080")
	if err != nil {
		panic(err)
	}
}
