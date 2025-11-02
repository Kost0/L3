package main

import (
	"github.com/Kost0/L3/internal/handlers"
	"github.com/Kost0/L3/internal/repository"
	"github.com/gin-contrib/cors"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()

	db, err := repository.ConnectDB()
	if err != nil {
		panic(err)
	}

	// Запускаем миграции
	err = repository.RunMigrations(db, "notifications")
	if err != nil {
		panic(err)
	}

	zlog.Logger.Info().Msg("DB started")

	handler := handlers.Handler{
		DB: db,
	}

	engine := ginext.New("")

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5000"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Accept", "Content-Type"},
	}))

	engine.POST("/comments", handler.CreateComment)

	engine.GET("/comments", handler.GetComments)

	engine.GET("/comments/all", handler.GetPageComments)

	engine.DELETE("/comments/:id", handler.DeleteComment)

	engine.GET("/comments/search", handler.SearchComment)

	err = engine.Run(":8080")
	if err != nil {
		panic(err)
	}
}
