package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	c_kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"go.uber.org/zap"

	"github.com/sibedge-llc/dp-services/eventer/internal/config"
	"github.com/sibedge-llc/dp-services/eventer/internal/event"
)

type Producer struct {
	ctx      context.Context
	producer *c_kafka.Producer
	topic    string
	id       uint64
	servers  string
}

func NewProducer(ctx context.Context, id uint64, cfg *config.KafkaConfig) (*Producer, error) {

	if cfg.Topic == "" {
		return nil, errors.New("topic name is empty or not provided")
	}

	p, err := c_kafka.NewProducer(&c_kafka.ConfigMap{
		"bootstrap.servers": cfg.BootstrapServers,
	})
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				zap.L().Info("stop listen send events.", zap.String("topic", cfg.Topic))
				return
			case e := <-p.Events():
				switch ev := e.(type) {
				case *c_kafka.Message:
					if ev.TopicPartition.Error != nil {
						zap.L().Error("Failed to deliver message", zap.Stringer("partition", ev.TopicPartition))
					} else {
						zap.L().Info("Successfully produced record", zap.Stringer("partition", ev.TopicPartition))
					}
				}
			}
		}
	}()

	return &Producer{
		ctx:      ctx,
		producer: p,
		topic:    cfg.Topic,
		id:       id,
		servers:  cfg.BootstrapServers,
	}, nil
}

func (p *Producer) Init(evt *event.Event) error {
	return p.ensureTopic(p.ctx, p.topic)
}

func (p *Producer) GetId() uint64 {
	return p.id
}

func (p *Producer) Send(evt *event.Event) error {
	return p.producer.Produce(&c_kafka.Message{
		TopicPartition: c_kafka.TopicPartition{Topic: &p.topic, Partition: c_kafka.PartitionAny},
		Key:            evt.Id,
		Value:          evt.Json,
	}, nil)
}

func (p *Producer) Flush() {
	p.producer.Flush(int(time.Second.Microseconds()))
}

func (p *Producer) Close() {
	p.producer.Close()
}

func (p *Producer) ensureTopic(ctx context.Context, topic string) error {
	a, err := c_kafka.NewAdminClientFromProducer(p.producer)
	if err != nil {
		return fmt.Errorf("Failed to create new admin client from producer: %s", err)
	}
	defer a.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	timeout := time.Second * 60
	results, err := a.CreateTopics(
		ctx,
		[]c_kafka.TopicSpecification{{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1}},
		c_kafka.SetAdminOperationTimeout(timeout))
	if err != nil {
		return fmt.Errorf("Admin Client request error: %v\n", err)
	}
	for _, result := range results {
		if result.Error.Code() != c_kafka.ErrNoError && result.Error.Code() != c_kafka.ErrTopicAlreadyExists {
			return fmt.Errorf("Failed to create topic: %v\n", result.Error)
		}
		zap.L().Info("topic create result", zap.Stringer("result", result))
	}
	return nil
}

func (p *Producer) GetServers() string {
	return p.servers
}
