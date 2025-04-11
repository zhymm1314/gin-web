// app/producer/abstract.go
package producer

import (
	"context"
	"fmt"
	"gin-web/config"
	"github.com/rabbitmq/amqp091-go"
)

type Producer interface {
	Publish(body []byte) error
	QueueName() string
}

type BaseProducer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	config  config.RabbitMQ
	queue   string
}

func NewBaseProducer(cfg config.RabbitMQ, queue string) (*BaseProducer, error) {
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

	return &BaseProducer{
		conn:    conn,
		channel: ch,
		config:  cfg,
		queue:   queue,
	}, err
}

func (p *BaseProducer) Publish(body []byte) error {
	return p.channel.PublishWithContext(
		context.Background(),
		"",      // exchange
		p.queue, // routing key
		false,   // mandatory
		false,   // immediate
		amqp091.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp091.Persistent,
		},
	)
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
