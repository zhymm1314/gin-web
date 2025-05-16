package consumer

import (
	"encoding/json"
	"gin-web/app/api"
	"github.com/streadway/amqp"
	"log"
)

type LogConsumer struct{}

func (c *LogConsumer) HandleMessage(msg amqp.Delivery) error {

	//return nil
	// 使用 defer + recover 捕获 panic 确保一定会ack处理
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
			//msg.Ack(false)
		}
	}()

	log.Printf("Processing order: %s", msg.Body)
	var raw map[string]interface{}
	if err := json.Unmarshal(msg.Body, &raw); err != nil {
		log.Printf("JSON解析失败: %v", err)
		//msg.Ack(false)
		return nil
	}

	// 提取 data 字段
	data, ok := raw["data"].(map[string]interface{})
	if !ok {
		log.Println("data字段不存在或类型错误")
		//msg.Ack(false)
		return nil
	}

	//global.App.Log.Info(string(msg.Body))
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
	api.SendTableStoreLog(params)

	//msg.Ack(false)
	return nil
}
