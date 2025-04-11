// app/producer/example.go
package producer

import (
	"gin-web/config"
	"gin-web/global"
)

type LogProducer struct {
	*BaseProducer
}

func NewLogProducer(cfg config.RabbitMQ) (*LogProducer, error) {
	base, err := NewBaseProducer(cfg, "log_queue")
	if err != nil {
		return nil, err
	}
	return &LogProducer{base}, nil
}

// 使用示例
func ExampleUsage() {
	cfg := global.App.Config.RabbitMQ
	p, _ := NewLogProducer(cfg)
	p.Publish([]byte(`{"level":"info","msg":"test"}`))
}
