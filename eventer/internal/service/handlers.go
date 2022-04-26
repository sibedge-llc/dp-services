package service

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sibedge-llc/dp-services/eventer/internal/event"
	"github.com/sibedge-llc/dp-services/eventer/internal/generator"
)

type GeneratorStatus struct {
	Id     string `json:"id"`
	Count  int64  `json:"count"`
	Active bool   `json:"active"`
}

type AddGeneratorResponse struct {
	Generators []GeneratorStatus `json:"generators"`
}

func (s *service) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	WriteObject(w, map[string]interface{}{
		"result": http.StatusText(http.StatusOK),
	})
}

func (s *service) handleGeneratorAdd(w http.ResponseWriter, r *http.Request) {
	var request event.GeneratorDesc
	err := ParseRequest(r, &request)
	if err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request: %v", err))
		return
	}

	if len(request.Destinations) == 0 {
		WriteError(w, http.StatusBadRequest, "destinations is not specified")
		return
	}

	if len(request.Events) == 0 {
		WriteError(w, http.StatusBadRequest, "events is not specified")
		return
	}

	eventDescs := make(map[string]event.EventDesc, len(request.Events))
	for _, eventDesc := range request.Events {
		if eventDesc.Id == "" {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("event id is empty or not defined"))
			return
		}
		eventDescs[eventDesc.Id] = eventDesc
	}

	destinations := make(map[string]generator.Destinaton, len(request.Destinations))
	for _, destination := range request.Destinations {
		if destination.Id == "" {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("destination id is empty or not defined"))
			return
		}
		var generatorDestination generator.Destinaton
		switch destination.Type {
		case event.DestinationTypeKafka:
			producer, err := s.kafkaService.Register(destination.Kafka)
			if err != nil {
				WriteError(w, http.StatusForbidden, fmt.Sprintf("failed to connect to kafka: %v", err))
				return
			}
			generatorDestination = producer
		case event.DestinationTypePostgres:
			db, err := s.postgresService.Register(destination.Postgres, time.Second)
			if err != nil {
				WriteError(w, http.StatusForbidden, fmt.Sprintf("failed to connect to kafka: %v", err))
				return
			}
			generatorDestination = db
		default:
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("unknown destination type: %v", destination.Type))
			return
		}
		destinations[destination.Id] = generatorDestination
	}

	var response AddGeneratorResponse
	for _, schedule := range request.Schedules {
		destination, ok := destinations[schedule.DestinationId]
		if !ok {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("destination with id = %s is not specified", schedule.DestinationId))
			return
		}
		eventDesc, ok := eventDescs[schedule.EventId]
		if !ok {
			WriteError(w, http.StatusBadRequest, fmt.Sprintf("event with id = %s is not specified", schedule.EventId))
			return
		}
		generator, err := s.generatorService.RegisterGenerator(eventDesc, destination)
		if err != nil {
			WriteError(w, http.StatusForbidden, fmt.Sprintf("failed to create generator: %v", err))
			return
		}
		count, isInfinite := generator.GetStatus()
		isActive := isInfinite || count > 0
		response.Generators = append(response.Generators, GeneratorStatus{
			Id:     generator.GetId(),
			Active: isActive,
			Count:  count,
		})
	}

	WriteObject(w, response)
}

func (s *service) handleGeneratorRemove(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Id string `json:"id"`
	}{}
	err := ParseRequest(r, &request)
	if err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request: %v", err))
		return
	}

	if request.Id == "" {
		WriteError(w, http.StatusBadRequest, "generator id is empty or not supplied")
		return
	}

	generatorId, err := strconv.ParseUint(request.Id, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request id %v: %v", request.Id, err))
		return
	}

	err = s.generatorService.UnregisterGenerator(generatorId)
	if err != nil && errors.Is(err, generator.ErrorNotFound) {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("generator with id is not found: %v", generatorId))
		return
	}
	WriteObject(w, nil)
}

func (s *service) handleGeneratorStatus(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Id string `json:"id"`
	}{}
	err := ParseRequest(r, &request)
	if err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request: %v", err))
		return
	}

	if request.Id == "" {
		WriteError(w, http.StatusBadRequest, "generator id is empty or not supplied")
		return
	}

	generatorId, err := strconv.ParseUint(request.Id, 10, 64)
	if err != nil {
		WriteError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse request id %v: %v", request.Id, err))
		return
	}

	generator := s.generatorService.GetGenerator(generatorId)
	if generator == nil {
		WriteError(w, http.StatusNotFound, fmt.Sprintf("generator with id is not found: %v", generatorId))
		return
	}
	count, isInfinite := generator.GetStatus()
	isActive := isInfinite || count > 0
	WriteObject(
		w,
		GeneratorStatus{
			Id:     generator.GetId(),
			Active: isActive,
			Count:  count,
		},
	)
}
