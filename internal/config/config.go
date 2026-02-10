package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/joho/godotenv"
)

type Role string

const (
	AdminRole  Role = "admin"
	UserRole   Role = "user"
	APIVersion      = "v2"
)

type App struct {
	Port                    string
	DatabaseURL             string
	LogLevel                slog.Level
	Environment             Environment
	AWS                     *AWS
	AuthProviderCredentials string
	ClientURLs              []string
	EmailService            *EmailService
	Metrics                 *Metrics
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
	CDNKeyName   string
}

type EmailService struct {
	SendingKey                   string
	Domain                       string
	Sender                       string
	Recipient                    string
	CourseCompletionTemplateName string
	CronSchedule                 string
}

var logLevelMap = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
	"warn":  slog.LevelWarn,
	"error": slog.LevelError,
}

type Metrics struct {
	Port string
}

func ParseEnv() (*App, error) {
	// Ignore error because in production there will be no .env file, env vars will be passed
	// in at runtime via docker run command/docker-compose
	_ = godotenv.Load()

	envVars := map[string]string{
		"SERVER_PORT":                     "",
		"DATABASE_URL":                    "",
		"LOG_LEVEL":                       "",
		"AWS_REGION":                      "",
		"AWS_ACCESS_KEY_ID":               "",
		"AWS_SECRET_ACCESS_KEY":           "",
		"AWS_BUCKET_NAME":                 "",
		"CLOUDFRONT_DOMAIN":               "",
		"CLOUDFRONT_KEY_PAIR_ID":          "",
		"CLOUDFRONT_KEY_NAME":             "",
		"ENVIRONMENT":                     "",
		"FIREBASE_CREDENTIALS":            "",
		"CLIENT_URLS":                     "",
		"MAILGUN_SENDING_KEY":             "",
		"MAILGUN_DOMAIN":                  "",
		"MAILGUN_SENDER":                  "",
		"MAILGUN_RECIPIENT":               "",
		"COURSE_COMPLETION_TEMPLATE_NAME": "",
		"EMAIL_FAILURE_CRON_SCHEDULE":     "",
		"METRICS_PORT":                    "",
	}

	for key := range envVars {
		value := os.Getenv(key)
		if value == "" {
			return nil, fmt.Errorf("%s environment variable is not set", key)
		}
		envVars[key] = value
	}

	logLevel, ok := logLevelMap[envVars["LOG_LEVEL"]]
	if !ok {
		return nil, errors.New("LOG_LEVEL should be one of debug|info|warning|error")
	}

	environment := Environment(envVars["ENVIRONMENT"])
	if !slices.Contains(validEnvironments, environment) {
		return nil, errors.New("ENVIRONMENT should be one of development|production|test")
	}

	clientURLsRaw := strings.Split(envVars["CLIENT_URLS"], ",")
	clientURLs := make([]string, 0, len(clientURLsRaw))
	for _, url := range clientURLsRaw {
		trimmed := strings.TrimSpace(url)
		if trimmed != "" {
			clientURLs = append(clientURLs, trimmed)
		}
	}

	return &App{
		Port:        envVars["SERVER_PORT"],
		DatabaseURL: envVars["DATABASE_URL"],
		LogLevel:    logLevel,
		AWS: &AWS{
			Region:       envVars["AWS_REGION"],
			AccessKey:    envVars["AWS_ACCESS_KEY_ID"],
			SecretKey:    envVars["AWS_SECRET_ACCESS_KEY"],
			BucketName:   envVars["AWS_BUCKET_NAME"],
			CDNDomain:    envVars["CLOUDFRONT_DOMAIN"],
			CDNKeyPairID: envVars["CLOUDFRONT_KEY_PAIR_ID"],
			CDNKeyName:   envVars["CLOUDFRONT_KEY_NAME"],
		},
		Environment:             environment,
		AuthProviderCredentials: envVars["FIREBASE_CREDENTIALS"],
		ClientURLs:              clientURLs,
		EmailService: &EmailService{
			SendingKey:                   envVars["MAILGUN_SENDING_KEY"],
			Domain:                       envVars["MAILGUN_DOMAIN"],
			CourseCompletionTemplateName: envVars["COURSE_COMPLETION_TEMPLATE_NAME"],
			CronSchedule:                 envVars["EMAIL_FAILURE_CRON_SCHEDULE"],
			Sender:                       envVars["MAILGUN_SENDER"],
			Recipient:                    envVars["MAILGUN_RECIPIENT"],
		},
		Metrics: &Metrics{
			Port: envVars["METRICS_PORT"],
		},
	}, nil
}
