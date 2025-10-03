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
	}
	if err != nil {
		panic(err)
	}
	
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	exc := rabbitmq.NewExchange("notifications", "topic")

	err = exc.BindToChannel(ch)
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

	err = ch.QueueBind(queueName, "#", exc.Name(), false, nil)
	if err != nil {
		panic(err)
	}

	publisher := rabbitmq.NewPublisher(ch, "")

	return publisher, manager, ch, conn
}
