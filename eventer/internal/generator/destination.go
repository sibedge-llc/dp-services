package generator

import (
	"github.com/sibedge-llc/dp-services/eventer/internal/event"
)

type Destinaton interface {
	Init(evt *event.Event) error
	GetId() uint64
	Send(evt *event.Event) error
	Flush()
	Close()
}
