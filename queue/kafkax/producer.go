package kafkax

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var producerPool sync.Map // key: topic, value: *KafkaProducer

type KafkaConfig struct {
	Username string
	Password string
	GroupID  string
	Brokers  string
}

type KafkaProducer struct {
	conn   *kafka.Conn
	config *KafkaConfig
	topic  string
}

// InitProducerForTopics 初始化每个 topic 的 producer
func InitProducerForTopics(ctx context.Context, c *KafkaConfig, topics []string) {
	for _, topic := range topics {
		producer := newKafkaProducerWithTopic(ctx, c, topic)
		producerPool.LoadOrStore(topic, producer)
		fmt.Println("Kafka producer initialized for topic:", topic)
	}
}

// 创建带 topic 的 producer
func newKafkaProducerWithTopic(ctx context.Context, c *KafkaConfig, topic string) *KafkaProducer {
	mechanism := plain.Mechanism{
		Username: c.Username,
		Password: c.Password,
	}
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
		KeepAlive:     10 * time.Second,
	}
	kConn, err := dialer.DialLeader(
		ctx,
		"tcp",
		c.Brokers,
		topic,
		0,
	)
	if err != nil {
		panic(err)
	}
	return &KafkaProducer{
		conn:   kConn,
		config: c,
		topic:  topic,
	}
}

// GetProducerByTopic 获取指定 topic 的 producer
func GetProducerByTopic(topic string) (*KafkaProducer, error) {
	val, ok := producerPool.Load(topic)
	if !ok {
		return nil, errors.New("producer not found for topic: " + topic)
	}
	return val.(*KafkaProducer), nil
}

// Publish 发布消息，自动重连
func (k *KafkaProducer) Publish(ctx context.Context, msg []kafka.Message) error {
	if err := k.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}
	_, err := k.conn.WriteMessages(msg...)
	if err != nil {
		// 检查是否为连接类错误
		if errors.Is(err, net.ErrClosed) || errors.Is(err, kafka.LeaderNotAvailable) {
			// 尝试重连
			if reconnectErr := k.reconnect(ctx); reconnectErr != nil {
				return reconnectErr
			}
			// 重连后重试
			_, err = k.conn.WriteMessages(msg...)
		}
	}
	return err
}

// 自动重连
func (k *KafkaProducer) reconnect(ctx context.Context) error {
	if k.conn != nil {
		_ = k.conn.Close()
	}
	mechanism := plain.Mechanism{
		Username: k.config.Username,
		Password: k.config.Password,
	}
	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
		KeepAlive:     10 * time.Second,
	}
	kConn, err := dialer.DialLeader(
		ctx,
		"tcp",
		k.config.Brokers,
		k.topic, 0)
	if err != nil {
		return err
	}
	k.conn = kConn
	return nil
}

// Close 关闭连接
func (k *KafkaProducer) Close() {
	if k.conn != nil {
		_ = k.conn.Close()
	}
}

// CloseAllProducers 关闭所有 producer
func CloseAllProducers() {
	producerPool.Range(func(key, value interface{}) bool {
		if p, ok := value.(*KafkaProducer); ok {
			p.Close()
		}
		return true
	})
}
