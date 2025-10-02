package rabbitMQ

import (
	"log"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
)

func StartConsumer(ch *rabbitmq.Channel) chan []byte {
	conn, err := rabbitmq.Connect("amqp://guest:guest@localhost:5672/", 3, time.Second)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()

	queueName := "allNotifications"

	err = ch.QueueBind(queueName, "#", "notifications", false, nil)
	if err != nil {
		log.Printf("error: %v", err)
		return nil
	}

	consumerCfg := rabbitmq.NewConsumerConfig(queueName)

	consumer := rabbitmq.NewConsumer(ch, consumerCfg)

	msgs := make(chan []byte)

	err = consumer.Consume(msgs)
	if err != nil {
		log.Printf("error: %v", err)
		return nil
	}

	return msgs
}
