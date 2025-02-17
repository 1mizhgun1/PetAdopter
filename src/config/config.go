package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type LoggerKey string
type UserIDKey string
type UsernameKey string

const (
	LoggerContextKey   LoggerKey = "logger"
	UserIDContextKey   UserIDKey = "userID"
	UsernameContextKey UserIDKey = "username"
)

type Config struct {
	Main       MainConfig       `yaml:"main"`
	Session    SessionConfig    `yaml:"session"`
	Validation ValidationConfig `yaml:"validation"`
	Ad         AdConfig         `yaml:"ad"`
}

type MainConfig struct {
	Port              string        `yaml:"port"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
	ShutdownTimeout   time.Duration `yaml:"shutdown_timeout"`
}

type SessionConfig struct {
	RefreshTokenLength    int           `yaml:"refresh_token_length"`
	AccessTokenLength     int           `yaml:"access_token_length"`
	AccessTokenLifeTime   time.Duration `yaml:"access_token_life_time"`
	AccessTokenCookieName string        `yaml:"access_token_cookie_name"`
	ProtectedCookies      bool          `yaml:"protected_cookies"`
}

type ValidationConfig struct {
	UsernameMinLength int `yaml:"username_min_length"`
	UsernameMaxLength int `yaml:"username_max_length"`
	PasswordMinLength int `yaml:"password_min_length"`
	PasswordMaxLength int `yaml:"password_max_length"`
}

type AdConfig struct {
	MaxPrice            int `yaml:"max_price"`
	DefaultSearchLimit  int `yaml:"default_search_limit"`
	DefaultSearchOffset int `yaml:"default_search_offset"`
}

func MustLoadConfig(path string, logger *slog.Logger) *Config {
	cfg := &Config{}

	file, err := os.Open(path)
	if err != nil {
		logger.Error(fmt.Sprintf("failed to open config file: %v", err))
		return &Config{}
	}
	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(cfg); err != nil {
		logger.Error(fmt.Sprintf("failed to decode config file: %v", err))
		return &Config{}
	}

	logger.Info("Successfully loaded config")
	return cfg
}
