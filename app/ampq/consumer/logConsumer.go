// app/consumer/email.go
package consumer

import (
	"gin-web/config"
	"github.com/rabbitmq/amqp091-go"
	"log"
)

type LogConsumer struct {
	*BaseConsumer
}

func (c *LogConsumer) QueueName() string {
	//TODO implement me
	panic("implement me")
}

func NewLogConsumer(cfg config.RabbitMQ) (Consumer, error) {
	base, err := NewBaseConsumer(cfg, "order_queue")
	if err != nil {
		return nil, err
	}
	return &LogConsumer{base}, nil
}

func (c *LogConsumer) Start() error {
	return c.BaseConsumer.Start(func(msg amqp091.Delivery) {
		log.Printf("收到日志消息: %s", msg.Body)

		//todo
	})
}
