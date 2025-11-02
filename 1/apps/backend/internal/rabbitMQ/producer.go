package rabbitMQ

import (
	"fmt"
	"os"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/zlog"
)

func InitRabbitMQ() (*rabbitmq.Publisher, *rabbitmq.QueueManager, *rabbitmq.Channel, *rabbitmq.Connection) {
	dsn := fmt.Sprintf("amqp://%s:%s@rabbitmq:5672",
		os.Getenv("RABBITMQ_DEFAULT_USER"),
		os.Getenv("RABBITMQ_DEFAULT_PASS"))

	zlog.Logger.Info().Msgf("dsn: %s", dsn)

	var conn *rabbitmq.Connection
	var err error

	for i := 0; i < 30; i++ {
		conn, err = rabbitmq.Connect(dsn, 3, time.Second)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	mainExchange := rabbitmq.NewExchange("notification-exchange", "topic")

	err = mainExchange.BindToChannel(ch)
	if err != nil {
		panic(err)
	}

	manager := rabbitmq.NewQueueManager(ch)

	queueConfig := rabbitmq.QueueConfig{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args:       nil,
	}

	queueName := "allNotifications"

	_, err = manager.DeclareQueue(queueName, queueConfig)
	if err != nil {
		panic(err)
	}

	err = ch.QueueBind(queueName, "#", mainExchange.Name(), false, nil)
	if err != nil {
		panic(err)
	}

	publisher := rabbitmq.NewPublisher(ch, "")

	return publisher, manager, ch, conn
}
