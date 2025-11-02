package main

import (
	"net/http"

	"github.com/Kost0/L3/internal/handlers"
	"github.com/Kost0/L3/internal/rabbitMQ"
	"github.com/Kost0/L3/internal/repository"
	"github.com/Kost0/L3/internal/sender"
	"github.com/Kost0/L3/internal/startRedis"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/zlog"
)

func main() {
	zlog.Init()
	// Запускаем соединение с RabbitMQ
	publisher, manager, ch, conn := rabbitMQ.InitRabbitMQ()

	zlog.Logger.Info().Msg("Connecting to RabbitMQ")

	defer conn.Close()

	messageChan := rabbitMQ.StartConsumer(ch)

	zlog.Logger.Info().Msg("Consumer started")

	sender.SendNotification(messageChan)

	zlog.Logger.Info().Msg("RabbitMQ started")

	// Запускаем соединение с базой данных
	db, err := repository.ConnectDB()
	if err != nil {
		panic(err)
	}

	// Запускаем мирации
	err = repository.RunMigrations(db, "notifications")
	if err != nil {
		panic(err)
	}

	err = repository.CheckMigrations(db)
	if err != nil {
		panic(err)
	}

	zlog.Logger.Info().Msg("DB started")

	// Запускаем соединение с Redis
	client := startRedis.StartRedis()

	// Структура для удобной работы handlers
	handler := handlers.Handler{
		Publisher:   publisher,
		Manager:     manager,
		DB:          db,
		RedisClient: client,
	}

	zlog.Logger.Info().Msg("Redis started")

	// Запускаем сервер
	engine := ginext.New()

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5000"},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Accept", "Content-Type"},
	}))

	engine.GET("notify/:id", handler.GetNotify)

	engine.POST("notify", handler.CreateNotify)

	engine.DELETE("notify/:id", handler.DeleteNotify)

	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	err = engine.Run(":8080")
	if err != nil {
		panic(err)
	}
}
