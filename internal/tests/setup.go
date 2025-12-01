package tests

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/compose"

	"github.com/supanova-rp/supanova-server/internal/app"
	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers/mocks"
)

const (
	testPort                = "3001"
	appStartTimeout         = 3 * time.Second
	readyCheckRetryInterval = 250 * time.Millisecond
)

type TestResources struct {
	ComposeStack *compose.DockerCompose
	DB           *sql.DB
	AppURL       string
}

// setupTestResources creates and starts all required containers for testing
func setupTestResources(ctx context.Context) (*TestResources, error) {
	composeStack, err := compose.NewDockerCompose("./docker-compose.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to create compose stack: %w", err)
	}

	err = composeStack.Up(ctx, compose.Wait(true))
	if err != nil {
		return nil, fmt.Errorf("failed to start compose stack: %w", err)
	}
	defer func() {
		// handle cleanup here if setup fails halfway through
		if err != nil {
			cleanupErr := composeStack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
			slog.Error("cleanup error", slog.Any("error", cleanupErr))
		}
	}()

	postgresURL, err := getPostgresURL(ctx, composeStack)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	mockObjectStorage := &mocks.ObjectStorageMock{
		GenerateUploadURLFunc: func(ctx context.Context, key string, contentType *string) (string, error) {
			return "https://mock-upload-url.com/" + key, nil
		},
		GetCDNURLFunc: func(ctx context.Context, key string) (string, error) {
			return "https://mock-cdn-url.com/" + key, nil
		},
	}

	cfg := &config.App{
		Port:        testPort,
		DatabaseURL: postgresURL,
		Environment: config.EnvironmentTest,
		AWS:         nil, // Not needed for tests with mock object storage
	}
	appURL := fmt.Sprintf("http://localhost:%s", testPort)

	// Start the app in a goroutine
	go func() {
		err := app.Run(ctx, cfg, app.Dependencies{
			ObjectStorage: mockObjectStorage,
		})
		if err != nil {
			slog.Error("app error", slog.Any("error", err))
		}
	}()

	err = waitForAppHealthy(ctx, appURL, appStartTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to start app: %v", err)
	}

	err = insertSetupData(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to insert test user data: %v", err)
	}

	return &TestResources{
		ComposeStack: composeStack,
		DB:           db,
		AppURL:       appURL,
	}, nil
}

func (tr *TestResources) Cleanup(ctx context.Context) {
	if tr == nil {
		return
	}

	if tr.DB != nil {
		err := tr.DB.Close()
		if err != nil {
			fmt.Printf("failed to close db: %v\n", err)
		}
	}

	if tr.ComposeStack != nil {
		err := tr.ComposeStack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
		if err != nil {
			fmt.Printf("failed to tear down compose stack: %v\n", err)
		}
	}
}

func getPostgresURL(ctx context.Context, composeStack *compose.DockerCompose) (string, error) {
	postgresContainer, err := composeStack.ServiceContainer(ctx, "postgres")
	if err != nil {
		return "", fmt.Errorf("failed to get postgres container: %w", err)
	}

	postgresPort, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		return "", fmt.Errorf("failed to get postgres mapped port: %w", err)
	}

	postgresHost, err := postgresContainer.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get postgres host: %w", err)
	}

	return fmt.Sprintf(
		"postgres://testuser:password@%s:%s/testdb?sslmode=disable",
		postgresHost,
		postgresPort.Port(),
	), nil
}

func insertSetupData(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(
		ctx,
		"INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
		testUserID,
		testUserName,
		testUserEmail,
	)

	return err
}

// waitForAppHealthy calls the /health endpoint until it gets a 200
// response or until the context is cancelled or the timeout is reached.
func waitForAppHealthy(
	parentCtx context.Context,
	appURL string,
	timeout time.Duration,
) error {
	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	client := http.Client{}

	for {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet,
			fmt.Sprintf("%s/%s/health", appURL, config.APIVersion),
			http.NoBody,
		)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("error making request: %s\n", err.Error())
			continue
		}
		if resp.StatusCode == http.StatusOK {
			resp.Body.Close() //nolint
			return nil
		}
		resp.Body.Close() //nolint

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(readyCheckRetryInterval):
			// retry
		}
	}
}
