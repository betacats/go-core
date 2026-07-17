package kafkax

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var producerPool sync.Map // key: name:topic, value: *KafkaProducer

type KafkaConfig struct {
	Username  string
	Password  string
	GroupID   string
	Brokers   string
	EnableTLS bool   // 是否启用 TLS，默认 false（不启用）
	Name      string // 实例标识，用于区分多个 Kafka 实例，为空时兼容旧行为；如果项目有多个 Kafka 实例，请务必设置！
}

// getPoolKey 构造 producerPool 的 key
func getPoolKey(kafkaName, topic string) string {
	if kafkaName == "" {
		return topic // 向后兼容：未设置 Name 时退化为纯 topic
	}
	return kafkaName + ":" + topic
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
		key := getPoolKey(c.Name, topic)
		producerPool.LoadOrStore(key, producer)
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
	if c.EnableTLS {
		dialer.TLS = &tls.Config{InsecureSkipVerify: true}
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
func GetProducerByTopic(kafkaName, topic string) (*KafkaProducer, error) {
	key := getPoolKey(kafkaName, topic)
	val, ok := producerPool.Load(key)
	if !ok {
		return nil, errors.New("producer not found for topic: " + topic)
	}
	return val.(*KafkaProducer), nil
}

// 判断是否为需要重连的连接错误
func isConnectionError(err error) bool {
	return errors.Is(err, net.ErrClosed) ||
		errors.Is(err, kafka.LeaderNotAvailable) ||
		errors.Is(err, io.EOF) ||
		errors.Is(err, syscall.EPIPE) ||
		errors.Is(err, syscall.ECONNRESET)
}

// Publish 发布消息，自动重连
func (k *KafkaProducer) Publish(ctx context.Context, msg []kafka.Message) error {
	var err error
	if err = k.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return err
	}
	_, err = k.conn.WriteMessages(msg...)
	if err != nil {
		// 扩展错误检查范围
		if isConnectionError(err) {
			// 尝试重连
			if reconnectErr := k.reconnect(ctx); reconnectErr != nil {
				return reconnectErr
			}
			// 重连后重试
			if err = k.conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
				return err
			}
			_, err = k.conn.WriteMessages(msg...)
		}
	}
	return err
}

// 自动重连
func (k *KafkaProducer) reconnect(ctx context.Context) error {
	// 触发重连了
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
	if k.config.EnableTLS {
		dialer.TLS = &tls.Config{InsecureSkipVerify: true}
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
	// 将重连后的生产者放回连接池
	producerPool.Store(getPoolKey(k.config.Name, k.topic), k)
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
