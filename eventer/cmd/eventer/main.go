package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kingpin"
	"github.com/sibedge-llc/dp-services/eventer/internal/config"
	"github.com/sibedge-llc/dp-services/eventer/internal/generator"
	"github.com/sibedge-llc/dp-services/eventer/internal/kafka"
	"github.com/sibedge-llc/dp-services/eventer/internal/postgres"
	"github.com/sibedge-llc/dp-services/eventer/internal/service"
	"go.uber.org/zap"
)

var (
	app          = kingpin.New("eventer", "Generate events on time interval basis.")
	commandStart = app.Command("start", "Start generate events.")

	startConfig = commandStart.Flag("config", "Config file.").Default("config.yaml").String()
)

func main() {
	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case commandStart.FullCommand():
		actionStart(*startConfig)
	}
}

func actionStart(configFile string) {
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		panic(err)
	}

	loggerConfig := zap.NewProductionConfig()
	err = loggerConfig.Level.UnmarshalText([]byte(cfg.Logging.Level))
	if err != nil {
		panic(err)
	}

	logger, err := loggerConfig.Build()
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaService, err := kafka.New(ctx, &cfg.Kafka)

	if err != nil {
		zap.L().Panic("create kafka service failed", zap.Error(err))
		return
	}

	postgresService, err := postgres.New(ctx, &cfg.Postgres)
	if err != nil {
		zap.L().Panic("create postgres service failed", zap.Error(err))
		return
	}

	generatorService, err := generator.New(ctx, cfg.InstanceId)
	if err != nil {
		zap.L().Panic("create generator service failed", zap.Error(err))
		return
	}

	service := service.New(&cfg.Service, kafkaService, postgresService, generatorService)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := service.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			zap.L().Panic("failed to run service", zap.Error(err))
			return
		}
	}()

	<-ch
}
