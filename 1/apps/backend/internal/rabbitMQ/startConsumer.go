package rabbitMQ

import (
	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

func StartConsumer(ch *rabbitmq.Channel) <-chan []byte {
	queueName := "allNotifications"

	consumerCfg := rabbitmq.NewConsumerConfig(queueName)

	consumer := rabbitmq.NewConsumer(ch, consumerCfg)

	msgs := make(chan []byte)

	go func() {
		defer close(msgs)
		err := consumer.Consume(msgs)
		if err != nil {
			zlog.Logger.Log().Msg(err.Error())
		}
	}()

	return msgs
}
