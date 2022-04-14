package generator

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/sibedge-llc/dp-services/eventer/internal/event"
	"github.com/sibedge-llc/dp-services/eventer/internal/utils"
)

var (
	ErrorNotFound = errors.New("not found")
)

type Service struct {
	ctx        context.Context
	instanceId string
	lock       sync.Mutex
	generators map[uint64]*Generator
}

func New(ctx context.Context, instanceId string) (*Service, error) {
	s := &Service{
		ctx:        ctx,
		instanceId: instanceId,
		generators: make(map[uint64]*Generator, 1),
	}
	return s, nil
}

func (s *Service) RegisterGenerator(eventDesc event.EventDesc, desination Destinaton) (*Generator, error) {
	eventId, err := utils.ObjectToJsonId(eventDesc, false)
	if err != nil {
		return nil, fmt.Errorf("failed to make id for event desc: %w", err)
	}

	generatorId, err := utils.ObjectToJsonId(map[string]interface{}{
		"event_id":       eventId,
		"destination_id": desination.GetId(),
	}, false)

	if err != nil {
		return nil, fmt.Errorf("failed to make id for generator: %w", err)
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	generator, ok := s.generators[generatorId]
	if ok && !generator.IsStopped() {
		return generator, nil
	}

	generator, err = NewGenerator(s.ctx, s.instanceId, generatorId, eventDesc, desination)
	if err != nil {
		return nil, fmt.Errorf("make generator failed: %w", err)
	}
	s.generators[generatorId] = generator

	return generator, nil
}

func (s *Service) UnregisterGenerator(generatorId uint64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	generator, ok := s.generators[generatorId]
	if ok {
		generator.Stop()
		delete(s.generators, generatorId)
		return nil
	}
	return ErrorNotFound
}

func (s *Service) GetGenerator(generatorId uint64) *Generator {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.generators[generatorId]
}
