package service

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/sibedge-llc/dp-services/eventer/internal/config"
	"github.com/sibedge-llc/dp-services/eventer/internal/generator"
	"github.com/sibedge-llc/dp-services/eventer/internal/kafka"
	"github.com/sibedge-llc/dp-services/eventer/internal/postgres"
)

type service struct {
	Listen           string
	generatorService *generator.Service
	kafkaService     *kafka.Service
	postgresService  *postgres.Service
}

func New(cfg *config.ServiceConfig, kafkaService *kafka.Service, postgresService *postgres.Service, generatorService *generator.Service) *service {
	return &service{
		Listen:           cfg.Listen,
		generatorService: generatorService,
		kafkaService:     kafkaService,
		postgresService:  postgresService,
	}
}

func (s *service) ListenAndServe() error {
	mux := mux.NewRouter()
	mux.HandleFunc("/generator/add", s.handleGeneratorAdd).Methods(http.MethodPost)
	mux.HandleFunc("/generator/remove", s.handleGeneratorRemove).Methods(http.MethodPost)
	mux.HandleFunc("/generator/status", s.handleGeneratorStatus).Methods(http.MethodGet)
	zap.L().Info("Server started", zap.String("listen", s.Listen))
	return http.ListenAndServe(s.Listen, mux)
}
