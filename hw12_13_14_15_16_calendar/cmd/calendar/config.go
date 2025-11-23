package main

import (
	"fmt"
	"os"

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
	Host string `toml:"host"`
	Port int    `toml:"port"`
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

	// defaults
	if cfg.Logger.Level == "" {
		cfg.Logger.Level = "info"
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Storage.Type == "" {
		cfg.Storage.Type = "memory"
	}

	return cfg, nil
}
