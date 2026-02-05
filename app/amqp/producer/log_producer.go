package producer

import (
	"gin-web/config"
)

// LogProducer 日志生产者
type LogProducer struct {
	*BaseProducer
}

// NewLogProducer 创建日志生产者实例
func NewLogProducer(cfg config.RabbitMQ) (*LogProducer, error) {
	base, err := NewBaseProducer(cfg, "log_queue")
	if err != nil {
		return nil, err
	}
	return &LogProducer{base}, nil
}

// 使用示例（通过依赖注入获取 cfg）:
// func ExampleUsage(cfg *config.Configuration) {
//     p, _ := NewLogProducer(cfg.RabbitMQ)
//     p.Publish([]byte(`{"level":"info","msg":"test"}`))
// }
