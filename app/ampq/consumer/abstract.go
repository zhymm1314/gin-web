// app/consumer/abstract.go
package consumer

import (
	"fmt"
	"gin-web/config"
	"github.com/rabbitmq/amqp091-go"
)

type Consumer interface {
	Start() error
	Stop()
	QueueName() string
}

type BaseConsumer struct {
	conn     *amqp091.Connection
	channel  *amqp091.Channel
	config   config.RabbitMQ
	queue    string
	stopChan chan struct{}
}

func NewBaseConsumer(cfg config.RabbitMQ, queue string) (*BaseConsumer, error) {
	conn, err := amqp091.Dial(getAMQPURI(cfg))
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	// 声明队列
	_, err = ch.QueueDeclare(
		queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)

	return &BaseConsumer{
		conn:     conn,
		channel:  ch,
		config:   cfg,
		queue:    queue,
		stopChan: make(chan struct{}),
	}, err
}

func (c *BaseConsumer) Start(handler func(amqp091.Delivery)) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg := <-msgs:
				handler(msg)
				msg.Ack(false)
			case <-c.stopChan:
				return
			}
		}
	}()
	return nil
}

func (c *BaseConsumer) Stop() {
	close(c.stopChan)
	c.channel.Close()
	c.conn.Close()
}

func getAMQPURI(cfg config.RabbitMQ) string {
	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Vhost,
	)
}
