package kafka

import (
	"time"

	"gin-web/pkg/mq"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// consumerGroupHandler 实现 sarama.ConsumerGroupHandler，
// 将 Kafka 投递转换为 mq.Message 交由业务 handler 处理。
type consumerGroupHandler struct {
	handler mq.Handler
	topic   string
	log     *zap.Logger
	done    chan struct{}
}

// Setup 在消费会话开始前调用。
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error { return nil }

// Cleanup 在消费会话结束后调用。
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim 消费一个分区的消息，直到会话结束。
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			mqMsg := &mq.Message{
				Body:      msg.Value,
				Topic:     msg.Topic,
				Headers:   toMQHeaders(msg.Headers),
				Partition: msg.Partition,
				Offset:    msg.Offset,
			}
			if err := h.handler.HandleMessage(mqMsg); err != nil {
				// 处理失败：不提交 offset，结束当前会话；
				// 下一会话从上次提交点重投（at-least-once），对应 RabbitMQ 的 Nack(requeue)。
				h.log.Error("handle message error",
					zap.String("topic", h.topic),
					zap.Int32("partition", msg.Partition),
					zap.Int64("offset", msg.Offset),
					zap.Error(err))
				// 退避后再重投，避免毒消息导致紧密循环
				select {
				case <-h.done:
				case <-time.After(reconnectBackoff):
				}
				return nil
			}
			// 处理成功：标记 offset，由 sarama 自动提交
			session.MarkMessage(msg, "")
		case <-h.done:
			return nil
		}
	}
}

// toMQHeaders 将 sarama.RecordHeader 切片转换为 mq.Message.Headers。
func toMQHeaders(headers []*sarama.RecordHeader) map[string][]byte {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string][]byte, len(headers))
	for _, h := range headers {
		out[string(h.Key)] = h.Value
	}
	return out
}
