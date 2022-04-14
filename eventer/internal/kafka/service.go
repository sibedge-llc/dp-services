package kafka

import (
	"context"
	"fmt"
	"sync"

	"github.com/sibedge-llc/dp-services/eventer/internal/config"
	"github.com/sibedge-llc/dp-services/eventer/internal/utils"
)

type Service struct {
	ctx       context.Context
	lock      sync.Mutex
	producers map[uint64]*Producer
	Default   *Producer
}

func New(ctx context.Context, cfg *config.KafkaConfig) (*Service, error) {
	s := &Service{
		ctx:       ctx,
		producers: make(map[uint64]*Producer, 1),
	}

	defaultProducer, err := s.Register(cfg)
	if err != nil {
		return nil, err
	}
	s.Default = defaultProducer

	return s, nil
}

func (s *Service) Register(cfg *config.KafkaConfig) (*Producer, error) {
	if cfg == nil {
		return s.Default, nil
	}

	id, err := utils.ObjectToJsonId(*cfg, false)
	if err != nil {
		return nil, fmt.Errorf("failed to make id for kafka config: %w", err)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	producer, ok := s.producers[id]
	if !ok {
		var err error
		if cfg.BootstrapServers == "" {
			cfg.BootstrapServers = s.Default.GetServers()
		}
		producer, err = NewProducer(s.ctx, id, cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create kafka producer: %w", err)
		}
		s.producers[id] = producer
	}
	return producer, nil
}
