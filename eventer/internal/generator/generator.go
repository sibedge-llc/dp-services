package generator

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/sibedge-llc/dp-services/eventer/internal/event"
	"go.uber.org/zap"
)

type Generator struct {
	id          uint64
	generator   *event.Generator
	destination Destinaton
	cancel      context.CancelFunc
	count       int64
	isInfinite  bool
}

func NewGenerator(
	ctx context.Context,
	instanceId string,
	generatorId uint64,
	eventDesc event.EventDesc,
	destination Destinaton,
) (*Generator, error) {
	name := fmt.Sprint(generatorId)
	zap.L().Debug("event", zap.String("id", name), zap.ByteString("schema", eventDesc.Schema))
	composer, err := event.NewComposerByContent(eventDesc.Dataset, instanceId, name, eventDesc.Schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create composed based on schema: %w", err)
	}

	interval, err := time.ParseDuration(eventDesc.Interval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time interval %v: %w", eventDesc.Interval, err)
	}

	if interval < time.Millisecond {
		return nil, fmt.Errorf("interval must be >= 1ms")
	}

	ctx, ctxCancel := context.WithCancel(ctx)
	stopped := make(chan struct{})
	cancel := func() {
		ctxCancel()
		<-stopped
	}
	if eventDesc.Count < -1 {
		eventDesc.Count = -1
	}

	s := &Generator{
		id:          generatorId,
		generator:   event.NewGenerator(ctx, interval, composer),
		destination: destination,
		cancel:      cancel,
		count:       eventDesc.Count,
		isInfinite:  eventDesc.Count <= 0,
	}

	evt := s.generator.Event(true)
	err = destination.Init(evt)
	if err != nil {
		return nil, fmt.Errorf("failed to init destination along event schema: %w", err)
	}

	go s.run(ctx, interval, stopped)

	return s, nil
}

func (s *Generator) GetId() string {
	return fmt.Sprint(s.id)
}

func (s *Generator) run(ctx context.Context, interval time.Duration, stopped chan struct{}) {
	ticker := time.NewTicker(interval)
	defer func() {
		ticker.Stop()
		s.destination.Flush()
		close(stopped)
		atomic.StoreInt64((*int64)(&s.count), -2)
	}()
	for s.Next() {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			evt := s.generator.Event()
			if !evt.IsValid() {
				if evt.IsStop {
					zap.L().Info("Stop event.")
					return
				}
				s.destination.Flush()
				zap.L().Info("No event.")
				continue
			}
			err := s.destination.Send(evt)
			if err != nil {
				zap.L().Error("send event to kafka failed.", zap.Error(err))
			}
		}
	}
}

func (s *Generator) GetStatus() (int64, bool) {
	count := atomic.LoadInt64((*int64)(&s.count))
	return count, s.isInfinite
}

func (s *Generator) Next() bool {
	count := atomic.LoadInt64((*int64)(&s.count))
	if count < -1 {
		return false
	}
	if s.isInfinite {
		return true
	}
	count = atomic.AddInt64((*int64)(&s.count), -1)
	return count >= 0
}

func (s *Generator) Stop() {
	s.cancel()
}

func (s *Generator) IsStopped() bool {
	count, isInfinite := s.GetStatus()
	if isInfinite && count >= -1 {
		return false
	}
	return count <= 0
}
