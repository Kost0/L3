package startKafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Kost0/L3/internal/photoProcessing"
	"github.com/Kost0/L3/internal/repository"
	"github.com/minio/minio-go/v7"
	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

func StartConsumer(ctx context.Context, db *dbpg.DB, client *minio.Client, topicName string) {
	brokerAddress := "kafka:9092"
	groupID := "myOrdersGroup-123456"

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          []string{brokerAddress},
		Topic:            topicName,
		GroupID:          groupID,
		MinBytes:         10e3,
		MaxBytes:         10e6,
		MaxWait:          1 * time.Second,
		RebalanceTimeout: 20 * time.Second,
		StartOffset:      kafka.FirstOffset,
		CommitInterval:   0,
	})
	defer func() {
		if err := reader.Close(); err != nil {
			log.Println(err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			zlog.Logger.Info().Msg("Shutting down kafka")
			return
		default:
			photo, err := getMessage(reader)
			if err != nil {
				zlog.Logger.Err(err)
				continue
			}

			err = photoProcessing.ProcessPhoto(client, photo, db)
			if err != nil {
				zlog.Logger.Err(err)
			}
		}
	}
}

func getMessage(reader *kafka.Reader) (*repository.Photo, error) {
	ctx := context.Background()

	msg, err := reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var photo repository.Photo

	err = json.Unmarshal(msg.Value, &photo)
	if err != nil {
		return nil, err
	}

	zlog.Logger.Info().Msg("Successfully processed message")

	return &photo, nil
}
