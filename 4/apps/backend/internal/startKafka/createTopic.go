package startKafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/zlog"
)

func EnsureTopicExists(brokerAddress, topicName string) error {
	ctx := context.Background()

	conn, err := kafka.DialLeader(ctx, "tcp", brokerAddress, topicName, 0)
	if err == nil {
		conn.Close()
		zlog.Logger.Info().Msgf("Topic %s is already exists", topicName)
		return nil
	}

	zlog.Logger.Info().Msgf("Creating topic %s", topicName)

	conn, err = kafka.Dial("tcp", brokerAddress)
	if err != nil {
		return err
	}

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = controllerConn.CreateTopics(topicConfigs...)
	if err != nil {
		return err
	}

	zlog.Logger.Info().Msgf("Creating topic %s success", topicName)
	return nil
}
