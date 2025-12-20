package main

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger    LoggerConf    `yaml:"logger"`
	DB        DBConf        `yaml:"db"`
	RabbitMQ  RabbitMQConf  `yaml:"rabbitmq"`
	Scheduler SchedulerConf `yaml:"scheduler"`
}

type LoggerConf struct {
	Level string `yaml:"level"`
}

type DBConf struct {
	DSN string `yaml:"dsn"`
}

type RabbitMQConf struct {
	URL       string `yaml:"url"`
	QueueName string `yaml:"queue_name"`
}

type SchedulerConf struct {
	Interval time.Duration `yaml:"interval"`
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
	if cfg.Scheduler.Interval == 0 {
		cfg.Scheduler.Interval = 1 * time.Minute
	}

	return cfg, nil
}
