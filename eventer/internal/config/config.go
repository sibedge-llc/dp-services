package config

import (
	"go.uber.org/config"
)

type Config struct {
	Logging    LoggingConfig  `yaml:"logging"`
	InstanceId string         `yaml:"instance_id"`
	Kafka      KafkaConfig    `yaml:"kafka"`
	Postgres   PostgresConfig `yaml:"postgres"`
	Service    ServiceConfig  `yaml:"service"`
}

type ServiceConfig struct {
	Listen string `yaml:"listen"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
}

type KafkaConfig struct {
	BootstrapServers string `yaml:"bootstrap_servers"`
	Topic            string `yaml:"topic"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Db       string `yaml:"db"`
	Table    string `yaml:"table"`
	Ssl      bool   `yaml:"ssl"`
}

func LoadConfig(configFile string) (*Config, error) {
	provider, err := config.NewYAML(config.File(configFile))
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = provider.Get(config.Root).Populate(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
