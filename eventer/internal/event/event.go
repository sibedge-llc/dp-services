package event

import (
	"github.com/sibedge-llc/dp-services/eventer/internal/config"
)

const (
	DestinationTypeKafka    = "kafka"
	DestinationTypePostgres = "postgres"
)

type EventDesc struct {
	Id       string `json:"id"`
	Dataset  string `json:"dataset"`
	Schema   []byte `json:"schema"`
	Count    int64  `json:"count,omitempty"`
	Interval string `json:"interval,omitempty"`
}

type DestinationDesc struct {
	Id       string                 `json:"id"`
	Type     string                 `json:"type"`
	Kafka    *config.KafkaConfig    `json:"kafka,omitempty"`
	Postgres *config.PostgresConfig `json:"postgres,omitempty"`
}

type ScheduleDesc struct {
	DestinationId string `json:"destination_id"`
	EventId       string `json:"event_id"`
}

type GeneratorDesc struct {
	Events       []EventDesc       `json:"events"`
	Destinations []DestinationDesc `json:"destinations"`
	Schedules    []ScheduleDesc    `json:"schedules"`
}
