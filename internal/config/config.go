package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RunMigrations bool
	LogLevel      slog.Level
	Environment   Environment
}

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
	EnvironmentTest        Environment = "test"
)

var validEnvironments = []Environment{
	EnvironmentDevelopment,
	EnvironmentProduction,
	EnvironmentTest,
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
		"ENVIRONMENT":    "",
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

	environment := Environment(envVars["ENVIRONMENT"])
	if !slices.Contains(validEnvironments, environment) {
		return nil, errors.New("ENVIRONMENT should be one of development|production|test")
	}

	return &Config{
		Port:          envVars["SERVER_PORT"],
		DatabaseURL:   envVars["DATABASE_URL"],
		RunMigrations: runMigrations,
		LogLevel:      logLevel,
		Environment:   environment,
	}, nil
}
