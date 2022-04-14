package event

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

type EventId []byte

var NoEventId EventId

type EventObject map[string]interface{}

type Event struct {
	Id     EventId
	Json   EventJson
	Object EventObject
	IsStop bool
}

var NoEvent = &Event{}
var StopEvent = &Event{IsStop: true}

func (evt *Event) IsValid() bool {
	return evt != nil && len(evt.Json) > 0
}

func (evt *Event) String() string {
	if evt == nil {
		return ""
	}
	return string(evt.Json)
}

type Generator struct {
	event    atomic.Value
	composer *Composer
}

func NewGenerator(ctx context.Context, interval time.Duration, composer *Composer) *Generator {
	generator := Generator{composer: composer}
	go generator.run(ctx, interval)
	return &generator
}

func (s *Generator) Event(forceUpdate ...bool) *Event {
	if len(forceUpdate) > 0 && forceUpdate[0] {
		err := s.generate()
		if err != nil {
			zap.L().Error("failed to update event", zap.Error(err))
			return StopEvent
		}
	}
	event, ok := s.event.Load().(*Event)
	if !ok {
		return NoEvent
	}
	defer s.event.Store(NoEvent)
	return event
}

func (s *Generator) run(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	err := s.generate()
	if err != nil {
		zap.L().Error("failed to generate event", zap.Error(err))
		return
	}
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			zap.L().Info("Generator interrupted")
			return
		case <-ticker.C:
			err := s.generate()
			if err != nil {
				zap.L().Error("failed to generate event", zap.Error(err))
				return
			}
		}
	}
}

func (s *Generator) generate() error {
	event, obj, err := s.composer.NewEvent()
	if err != nil {
		s.event.Store(StopEvent)
		return err
	}
	s.event.Store(&Event{Json: event, Id: GetId(obj), Object: obj})
	return nil
}

func GetId(v EventObject) EventId {
	id, ok := v["id"]
	if !ok {
		zap.L().Error("failed to detect id from event")
		return NoEventId
	}

	switch t := id.(type) {
	case nil:
		return NoEventId
	case string:
		return EventId(t)
	default:
		return EventId(fmt.Sprint(t))
	}
}
