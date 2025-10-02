package rabbitMQ

import (
	"time"

	"github.com/wb-go/wbf/rabbitmq"
)

func InitRabbitMQ() (*rabbitmq.Publisher, *rabbitmq.QueueManager, *rabbitmq.Channel) {
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

	ch := &rabbitmq.Channel{}

	manager := rabbitmq.NewQueueManager(ch)

	publisher := rabbitmq.NewPublisher(ch, "")

	return publisher, manager, ch
}
