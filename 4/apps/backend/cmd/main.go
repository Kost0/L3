package main

import (
	"context"

	"github.com/Kost0/L3/internal/handlers"
	"github.com/Kost0/L3/internal/minIO"
	"github.com/Kost0/L3/internal/repository"
	"github.com/Kost0/L3/internal/startKafka"
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

	err = repository.RunMigrations(db, "notifications")
	if err != nil {
		panic(err)
	}

	zlog.Logger.Info().Msg("DB started")

	client := minIO.InitMinio()

	zlog.Logger.Info().Msg("MinIO started")

	ctx := context.Background()

	topicName := "test1234"

	err = startKafka.EnsureTopicExists("kafka:9092", topicName)

	go startKafka.StartConsumer(ctx, db, client, topicName)

	writer := startKafka.StartProducer(topicName)

	defer writer.Close()

	handler := handlers.Handler{
		DB:     db,
		Client: client,
		Writer: writer,
	}

	engine := ginext.New("")

	engine.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5000",
			"http://172.22.0.7:5000",
		},
		AllowMethods: []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Accept", "Content-Type"},
	}))

	engine.POST("upload", handler.ProcessPhoto)

	engine.GET("image/:id", handler.GetPhoto)

	engine.DELETE("image/:id", handler.DeletePhoto)

	engine.GET("status/:id", handler.GetPhotoStatus)

	err = engine.Run(":8080")
	if err != nil {
		panic(err)
	}
}
