package main

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Logger  LoggerConf  `toml:"logger"`
	Server  ServerConf  `toml:"server"`
	Storage StorageConf `toml:"storage"`
	DB      DBConf      `toml:"db"`
}

type LoggerConf struct {
	Level string `toml:"level"`
}

type ServerConf struct {
	Host     string `yaml:"host"`
	HTTPPort int    `yaml:"http_port"`
	GRPCPort int    `yaml:"grpc_port"`
	// Для обратной совместимости
	Port int `yaml:"port"`
}

type StorageConf struct {
	Type string `toml:"type"` // "memory" или "sql"
}

type DBConf struct {
	DSN string `toml:"dsn"`
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
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("HTTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.HTTPPort = p
		}
	}
	if port := os.Getenv("GRPC_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.GRPCPort = p
		}
	}
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		cfg.Storage.Type = storageType
	}
	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		cfg.DB.DSN = dsn
	}

	// defaults
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	// Обратная совместимость: если указан старый port, используем его для HTTP
	if cfg.Server.Port != 0 && cfg.Server.HTTPPort == 0 {
		cfg.Server.HTTPPort = cfg.Server.Port
	}
	if cfg.Server.HTTPPort == 0 {
		cfg.Server.HTTPPort = 8080
	}
	if cfg.Server.GRPCPort == 0 {
		cfg.Server.GRPCPort = 50051
	}
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = "memory"
	}

	return cfg, nil
}
