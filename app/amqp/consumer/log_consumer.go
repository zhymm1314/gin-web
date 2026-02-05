package consumer

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	"gin-web/app/api"
	"gin-web/config"
)

// LogConsumer 日志消费者（通过依赖注入获取 config 和 logger）
type LogConsumer struct {
	cfg *config.Configuration
	log *zap.Logger
}

// NewLogConsumer 创建日志消费者实例
func NewLogConsumer(cfg *config.Configuration, log *zap.Logger) *LogConsumer {
	return &LogConsumer{cfg: cfg, log: log}
}

func (c *LogConsumer) HandleMessage(msg amqp.Delivery) error {
	// 使用 defer + recover 捕获 panic 确保一定会ack处理
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	log.Printf("Processing order: %s", msg.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(msg.Body, &raw); err != nil {
		log.Printf("JSON解析失败: %v", err)
		return nil
	}

	// 提取 data 字段
	data, ok := raw["data"].(map[string]interface{})
	if !ok {
		log.Println("data字段不存在或类型错误")
		return nil
	}

	params := api.LogParams{
		Data: struct {
			ReqID    string `json:"req_id"`
			Name     string `json:"name"`
			LogLevel int    `json:"log_level"`
			Detail   string `json:"detail"`
			Custom1  string `json:"custom1,omitempty"`
			Custom2  string `json:"custom2,omitempty"`
		}{
			ReqID:    data["req_id"].(string),
			Name:     data["type"].(string),
			LogLevel: 10,
			Detail:   string(msg.Body),
			Custom1:  "test",
			Custom2:  "test",
		},
		TableName: "eric_request_logs",
	}
	api.SendTableStoreLog(c.cfg, c.log, params)

	return nil
}
