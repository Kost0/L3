package main

import (
	"github.com/wb-go/wbf/ginext"
	"internal/handlers"
	"internal/rabbitMQ"
	"internal/repository"
	"internal/sender"
	"internal/startRedis"
)

func main() {
	// Запускаем соединение с RabbitMQ
	publisher, manager, ch := rabbitMQ.InitRabbitMQ()

	messageChan := rabbitMQ.StartConsumer(ch)

	handlers.
		sender.SendNotification(messageChan)

	// Запускаем соединение с базой данных
	db, err := repository.ConnectDB()
	if err != nil {
		panic(err)
	}

	// Запускаем мирации
	err = repository.RunMigrations(db, "postgres")
	if err != nil {
		panic(err)
	}

	// Запускаем соединение с Redis
	client := startRedis.StartRedis()

	// Структура для удобной работы handlers
	handler := handlers.Handler{
		Publisher:   publisher,
		Manager:     manager,
		DB:          db,
		RedisClient: client,
	}

	// Запускаем сервер
	engine := ginext.New()

	engine.GET("notify/:id", handler.GetNotify)

	engine.POST("notify", handler.CreateNotify)

	engine.DELETE("notify/:id", handler.DeleteNotify)

	err = engine.Run("localhost:8080")
	if err != nil {
		panic(err)
	}
}
