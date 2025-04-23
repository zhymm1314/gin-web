// app/consumer/abstract.go
package consumer

import (
	"fmt"
	"gin-web/config"
	"github.com/rabbitmq/amqp091-go"
	"log"
	"math"
	"net"
	"sync"
	"time"
)

const maxRetries = 10

type Consumer interface {
	Start() error
	Stop()
}

func getConsumerTypes(ct string, cfg config.RabbitMQ, queueName string) (Consumer, error) {
	if ct == "log" {
		return NewLogConsumer(cfg, queueName)
	}

	return nil, fmt.Errorf("consumer type %s not found", ct)
}

type BaseConsumer struct {
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	config    config.RabbitMQ
	queue     string
	stopChan  chan struct{}
	handler   func(amqp091.Delivery)
	mu        sync.Mutex
	isRunning bool // 新增运行状态标识
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
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isRunning {
		return nil // 避免重复启动
	}

	c.handler = handler

	msgs, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	c.isRunning = true
	go c.listen(msgs)
	return nil
}

func (c *BaseConsumer) listen(msgs <-chan amqp091.Delivery) {
	defer func() {
		c.mu.Lock()
		c.isRunning = false
		c.mu.Unlock()
	}()

	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				log.Println("消息通道异常关闭，触发重连")
				c.reconnect()
				return
			}
			c.handler(msg)
		case <-time.After(30 * time.Second):
			if c.conn.IsClosed() {
				log.Println("检测到连接异常，触发重连")
				c.reconnect()
				return
			}
		case <-c.stopChan:
			return
		}
	}
}

func (c *BaseConsumer) reconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 强制关闭旧连接
	if c.channel != nil {
		c.channel.Close()
		c.channel = nil
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	// 增加连接超时设置
	dialConfig := amqp091.Config{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 10*time.Second)
		},
	}

	for i := 0; i < maxRetries; i++ {
		wait := time.Duration(math.Pow(2, float64(i))) * time.Second
		log.Printf("等待 %.0f 秒后重试...", wait.Seconds())
		time.Sleep(wait)

		// 使用带超时的 Dial 方法
		conn, err := amqp091.DialConfig(getAMQPURI(c.config), dialConfig)
		if err != nil {
			log.Printf("第 %d 次连接失败: %v", i+1, err)
			continue
		}

		// 验证连接是否有效
		if conn.IsClosed() {
			conn.Close()
			continue
		}

		// 创建新通道
		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			log.Printf("通道创建失败: %v", err)
			continue
		}

		// 更新连接和通道
		c.conn = conn
		c.channel = ch
		log.Printf("RabbitMQ 重连成功")

		// 重新声明队列并启动消费者
		if _, err := ch.QueueDeclare(c.queue, true, false, false, false, nil); err != nil {
			log.Printf("队列声明失败: %v", err)
			continue
		}

		// 新增状态重置
		c.mu.Lock()
		c.isRunning = false
		c.mu.Unlock()

		// 修改启动方式
		if err := c.Start(c.handler); err != nil {
			return fmt.Errorf("重连后启动失败: %v", err)
		}
		return nil
	}
	return fmt.Errorf("超过最大重试次数")
}

func (c *BaseConsumer) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	close(c.stopChan)
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.isRunning = false
}

func getAMQPURI(cfg config.RabbitMQ) string {
	// 检查关键参数是否为空
	if cfg.Username == "" || cfg.Password == "" || cfg.Host == "" {
		log.Fatal("RabbitMQ 配置参数缺失")
	}

	fmt.Println(fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Vhost,
	))

	return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Vhost,
	)
}
