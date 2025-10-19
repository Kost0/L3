package startKafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Kost0/L3/internal/repository"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/zlog"
)

func StartProducer(topicName string) *kafka.Writer {
	writer := &kafka.Writer{
		Addr:     kafka.TCP("kafka:9092"),
		Topic:    topicName,
		Balancer: &kafka.LeastBytes{},
	}

	return writer
}

func SendMessage(photo *repository.Photo, writer *kafka.Writer) error {
	data, err := json.Marshal(photo)
	if err != nil {
		return err
	}

	ctx := context.Background()

	err = writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte("photo"),
		Value: data,
		Time:  time.Now(),
	})

	if err != nil {
		return err
	}

	zlog.Logger.Info().Msg("Successfully sent message")

	return nil
}
