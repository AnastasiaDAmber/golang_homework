package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger   LoggerConf   `yaml:"logger"`
	RabbitMQ RabbitMQConf `yaml:"rabbitmq"`
	DB       DBConf       `yaml:"db"`
}

type DBConf struct {
	DSN string `yaml:"dsn"`
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

	// Переопределяем из переменных окружения, если они заданы
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}
	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		cfg.RabbitMQ.URL = url
	}
	if queueName := os.Getenv("RABBITMQ_QUEUE_NAME"); queueName != "" {
		cfg.RabbitMQ.QueueName = queueName
	}
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		cfg.DB.DSN = dsn
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
