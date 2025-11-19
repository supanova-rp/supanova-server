package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RunMigrations bool
	LogLevel      slog.Level
}

var logLevelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func ParseEnv() (*Config, error) {
	// Ignore error because in production there will be no .env file, env vars will be passed
	// in at runtime via docker run command/docker-compose
	_ = godotenv.Load()

	envVars := map[string]string{
		"SERVER_PORT":    "",
		"DATABASE_URL":   "",
		"RUN_MIGRATIONS": "",
		"LOG_LEVEL":      "",
	}

	for key := range envVars {
		value := os.Getenv(key)
		if value == "" {
			return nil, fmt.Errorf("%s environment variable is not set", key)
		}
		envVars[key] = value
	}

	runMigrations, err := strconv.ParseBool(envVars["RUN_MIGRATIONS"])
	if err != nil {
		return nil, fmt.Errorf("unable to parse RUN_MIGRATIONS environment variable: %v", err)
	}

	logLevel, ok := logLevelMap[envVars["LOG_LEVEL"]]
	if !ok {
		return nil, errors.New("LOG_LEVEL should be one of debug|info|warning|error")
	}

	return &Config{
		Port:          envVars["SERVER_PORT"],
		DatabaseURL:   envVars["DATABASE_URL"],
		RunMigrations: runMigrations,
		LogLevel:      logLevel,
	}, nil
}
