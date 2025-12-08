package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger   LoggerConf   `yaml:"logger"`
	RabbitMQ RabbitMQConf `yaml:"rabbitmq"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type RabbitMQConf struct {
	URL       string `yaml:"url"`
	QueueName string `yaml:"queue_name"`
}

func NewConfigFromFile(path string) (Config, error) {
	if path == "" {
		return Config{}, fmt.Errorf("empty config path")
	}

	f, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return Config{}, err
	}

	// defaults
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.RabbitMQ.QueueName == "" {
		cfg.RabbitMQ.QueueName = "notifications"
	}

	return cfg, nil
}
