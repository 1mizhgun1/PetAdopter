package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type LoggerKey string

const LoggerContextKey LoggerKey = "logger"

type Config struct {
	Main MainConfig `yaml:"main"`
}

type MainConfig struct {
	Port              string        `yaml:"port"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}

func MustLoadConfig(path string, logger *slog.Logger) *Config {
	cfg := &Config{}

	file, err := os.Open(path)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to open config file: %v", err))
		return &Config{}
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err = d.Decode(cfg); err != nil {
		logger.Error(fmt.Sprintf("failed to decode config file: %v", err))
		return &Config{}
	}

	logger.Info("Successfully loaded config")
	return cfg
}
