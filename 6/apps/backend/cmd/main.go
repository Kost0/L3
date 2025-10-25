package main

import (
	"github.com/Kost0/L3/internal/handlers"
	"github.com/Kost0/L3/internal/repository"
	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

const (
	OrderRoute       = "items"
	OrderById        = "items/:id"
	AnalyticsRoute   = "analytics"
	CategoriesRoute  = "categories"
	OrdersByCategory = "orders/:category"
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

	engine := ginext.New("")

	handler := handlers.Handler{
		DB: db,
	}

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5000",
		},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Accept", "Content-Type"},
	}))

	engine.POST(OrderRoute, handler.CreateOrder)

	engine.GET(OrderRoute, handler.GetOrders)

	engine.PUT(OrderById, handler.UpdateOrder)

	engine.DELETE(OrderById, handler.DeleteOrder)

	engine.GET(AnalyticsRoute, handler.GetAnalytics)

	engine.GET(CategoriesRoute, handler.GetAllCategories)

	engine.GET(OrdersByCategory, handler.GetOrdersByCategory)

	err = engine.Run(":8080")
	if err != nil {
		panic(err)
	}
}
