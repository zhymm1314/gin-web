package bootstrap

import (
	"errors"
	"gin-web/app/amqp/consumer"
	"gin-web/config"
	"gin-web/global"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

type Consumer struct {
	queueName string
	handler   consumer.ConsumerHandler
	conn      *amqp.Connection
	done      chan struct{}
}

func NewConsumer(conn *amqp.Connection, queueName string, handler consumer.ConsumerHandler) *Consumer {
	return &Consumer{
		queueName: queueName,
		handler:   handler,
		conn:      conn,
		done:      make(chan struct{}),
	}
}

func (c *Consumer) Start() {
	for {
		select {
		case <-c.done:
			return
		default:
			err := c.consume()
			if err != nil {
				// 自动重试逻辑
				continue
			}
		}
	}
}

func (c *Consumer) consume() error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// 设置预取值为 5（每次最多处理 5 条未确认消息）
	err = ch.Qos(
		120,   // prefetchCount（预取值）
		0,     // prefetchSize（预取大小，通常为 0）
		false, // global（是否全局应用）
	)
	if err != nil {
		log.Fatalf("设置预取值失败: %v", err)
	}

	q, err := ch.QueueDeclare(
		c.queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name,
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

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				return errors.New("message channel closed")
			}
			if err := c.handler.HandleMessage(msg); err != nil {
				msg.Nack(false, true)
			} else {
				msg.Ack(false)
			}
		case <-c.done:
			return nil
		}
	}
}

func (c *Consumer) Stop() {
	close(c.done)
}

func InitRabbitmq() {

	// 加载配置
	cfgConsumer, err := config.LoadConfig("./config/yaml/consumer.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 注册消费者处理器
	handlers := map[string]consumer.ConsumerHandler{
		"LogConsumer": &consumer.LogConsumer{},
		//"PaymentConsumer": &consumer.PaymentConsumer{},
	}

	cfg := config.RabbitMQConfig{
		Host:     global.App.Config.RabbitMQ.Host,
		Port:     global.App.Config.RabbitMQ.Port,
		Username: global.App.Config.RabbitMQ.Username,
		Password: global.App.Config.RabbitMQ.Password,
		Vhost:    global.App.Config.RabbitMQ.Vhost,
	}

	appCfg := &config.AppConfig{
		Consumers: cfgConsumer.Consumers,
		RabbitMQ:  cfg,
	}
	// 创建消费者管理器
	cm := NewConsumerManager(appCfg, handlers)
	go cm.Start()

	// 初始化Gin路由（可选）
	// r := gin.Default()
	// r.GET("/health", func(c *gin.Context) {
	//     c.JSON(200, gin.H{"status": "ok"})
	// })
	// r.Run()

	// 保持主程序运行
	select {}
}
