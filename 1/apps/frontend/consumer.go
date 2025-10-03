package frontend

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
)

func StartConsumer(ch *rabbitmq.Channel) (chan []byte, *rabbitmq.Connection) {
	dsn := fmt.Sprintf("amqp://%s:%s@rabbitmq:5672/",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"))

	conn, err := rabbitmq.Connect(dsn, 3, time.Second)
	if err != nil {
		panic(err)
	}

	queueName := "allNotifications"

	err = ch.QueueBind(queueName, "#", "notifications", false, nil)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, nil
	}

	consumerCfg := rabbitmq.NewConsumerConfig(queueName)

	consumer := rabbitmq.NewConsumer(ch, consumerCfg)

	msgs := make(chan []byte)

	err = consumer.Consume(msgs)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, nil
	}

	return msgs, conn
}
