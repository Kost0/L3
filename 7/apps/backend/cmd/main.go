package main

import (
	"github.com/Kost0/L3/internal/handlers"
	"github.com/Kost0/L3/internal/middleware"
	"github.com/Kost0/L3/internal/repository"
	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

const (
	itemsRoute = "/items"
	itemById   = ":id"
	login      = "login"
	history    = "history/:id"
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
		AllowMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Accept", "Content-Type", "Authorization"},
	}))

	engine.POST(login, handler.Login)

	api := engine.Group(itemsRoute)
	api.Use(middleware.AuthMiddleware())
	{
		api.POST("", middleware.RoleCheckMiddleware("admin", "manager"), handler.CreateItem)

		api.GET("", middleware.RoleCheckMiddleware("admin", "manager", "viewer"), handler.GetItems)

		api.PUT(itemById, middleware.RoleCheckMiddleware("admin", "manager"), handler.UpdateItem)

		api.DELETE(itemById, middleware.RoleCheckMiddleware("admin"), handler.DeleteItem)

		api.GET(history, middleware.RoleCheckMiddleware("admin", "manager", "viewer"), handler.GetHistory)
	}

	err = engine.Run(":8080")
	if err != nil {
		panic(err)
	}
}
