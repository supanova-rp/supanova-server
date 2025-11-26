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

type App struct {
	Port          string
	DatabaseURL   string
	RunMigrations bool
	LogLevel      slog.Level
	Environment   Environment
	AWS           *AWS
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

type AWS struct {
	Region       string
	AccessKey    string
	SecretKey    string
	BucketName   string
	CDNDomain    string
	CDNKeyPairID string
}

var logLevelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

func ParseEnv() (*App, error) {
	// Ignore error because in production there will be no .env file, env vars will be passed
	// in at runtime via docker run command/docker-compose
	_ = godotenv.Load()

	envVars := map[string]string{
		"SERVER_PORT":            "",
		"DATABASE_URL":           "",
		"RUN_MIGRATIONS":         "",
		"LOG_LEVEL":              "",
		"AWS_REGION":             "",
		"AWS_ACCESS_KEY_ID":      "",
		"AWS_SECRET_ACCESS_KEY":  "",
		"AWS_BUCKET_NAME":        "",
		"CLOUDFRONT_DOMAIN":      "",
		"CLOUDFRONT_KEY_PAIR_ID": "",
		"ENVIRONMENT":            "",
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

	return &App{
		Port:          envVars["SERVER_PORT"],
		DatabaseURL:   envVars["DATABASE_URL"],
		RunMigrations: runMigrations,
		LogLevel:      logLevel,
		AWS: &AWS{
			Region:       envVars["AWS_REGION"],
			AccessKey:    envVars["AWS_ACCESS_KEY_ID"],
			SecretKey:    envVars["AWS_SECRET_ACCESS_KEY"],
			BucketName:   envVars["AWS_BUCKET_NAME"],
			CDNDomain:    envVars["CLOUDFRONT_DOMAIN"],
			CDNKeyPairID: envVars["CLOUDFRONT_KEY_PAIR_ID"],
		},
		Environment: environment,
	}, nil
}
