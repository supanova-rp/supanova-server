package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/testcontainers/testcontainers-go/modules/compose"
)

type TestResources struct {
	ComposeStack *compose.DockerCompose
	DB           *sql.DB
	AppURL       string
}

// setupTestResources creates and starts all required containers for testing
func setupTestResources(ctx context.Context, t *testing.T) (*TestResources, error) {
	composeStack, err := compose.NewDockerCompose("./docker-compose.yml")
	if err != nil {
		return nil, fmt.Errorf("failed to create compose stack: %w", err)
	}

	err = composeStack.Up(ctx, compose.Wait(true))
	if err != nil {
		return nil, fmt.Errorf("failed to start compose stack: %w", err)
	}

	postgresURL, err := getPostgresURL(ctx, composeStack)
	if err != nil {
		return nil, err
	}

	appURL, err := getAppURL(ctx, composeStack)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	return &TestResources{
		ComposeStack: composeStack,
		DB:           db,
		AppURL:       appURL,
	}, nil
}

func (tr *TestResources) Cleanup(ctx context.Context, t *testing.T) {
	if tr.DB != nil {
		err := tr.DB.Close()
		if err != nil {
			t.Logf("failed to close db: %v", err)
		}
	}

	if tr.ComposeStack != nil {
		fmt.Println(">>> downnnnn")
		err := tr.ComposeStack.Down(ctx, compose.RemoveOrphans(true), compose.RemoveImagesLocal)
		if err != nil {
			t.Logf("failed to tear down compose stack: %v", err)
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

func getAppURL(ctx context.Context, composeStack *compose.DockerCompose) (string, error) {
	appContainer, err := composeStack.ServiceContainer(ctx, "go-template")
	if err != nil {
		return "", fmt.Errorf("failed to get app container: %w", err)
	}

	appPort, err := appContainer.MappedPort(ctx, "3001")
	if err != nil {
		return "", fmt.Errorf("failed to get app mapped port: %w", err)
	}

	appHost, err := appContainer.Host(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get app host: %w", err)
	}

	return fmt.Sprintf("http://%s:%s", appHost, appPort.Port()), nil
}
