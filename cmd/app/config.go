package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

// Config ...
type Config struct {
	BindAddr   string `env:"BIND_ADDR,default=:8080"`
	LogLevel   string `env:"LOG_LEVEL,default=info"`
	TGAppIDStr string `env:"TG_APP_ID"`
	TGAppID    int
	TGAppHash  string `env:"TG_APP_HASH"`
}

// ParseConfig загружает .env файл (если есть) и парсит переменные окружения.
func parseConfig() (Config, error) {
	if err := godotenv.Load("config.env"); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load env variables from file: %w", err)
	}

	var c Config
	if err := envconfig.Process(context.Background(), &c); err != nil {
		return Config{}, fmt.Errorf("parse env variables to config: %w", err)
	}

	if c.BindAddr == "" {
		return Config{}, errors.New("var BIND_ADDRESS is required")
	}

	if c.TGAppIDStr == "" {
		return Config{}, errors.New("var TG_APP_ID is required")
	}

	appID, err := strconv.Atoi(c.TGAppIDStr)

	if err != nil {
		return Config{}, errors.New("var TG_APP_ID convert err")
	}

	c.TGAppID = appID

	if c.TGAppHash == "" {
		return Config{}, errors.New("var TG_APP_HASH is required")
	}

	return c, nil
}
