package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/JDGarner/go-template/internal/store/sqlc"
)

func TestIntegration(t *testing.T) {
	ctx := context.Background()

	testResources, err := setupTestResources(ctx, t)
	if err != nil {
		fmt.Printf("setup tests failed: %s", err)
		testResources.Cleanup(ctx, t)
		os.Exit(1)
	}

	t.Cleanup(func() {
		testResources.Cleanup(ctx, t)
	})

	t.Run("returns dummy item", func(t *testing.T) {
		id := uuid.New()
		expectedName := "hello"

		_, err := testResources.DB.ExecContext(
			ctx,
			`INSERT INTO dummy VALUES ($1, $2)`,
			id,
			expectedName,
		)
		if err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		resp := getItem(t, testResources.AppURL, id)
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		var result sqlc.Dummy
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatalf("failed to parse JSON response: %v. Body: %s", err, string(body))
		}

		if result.Name != expectedName {
			t.Errorf("expected name '%s', got '%s'", expectedName, result.Name)
		}

		if result.ID.String() != id.String() {
			t.Errorf("expected id '%s', got '%s'", id.String(), result.ID)
		}
	})

	t.Run("returns not found error", func(t *testing.T) {
		nonExistantID := uuid.New()

		resp := getItem(t, testResources.AppURL, nonExistantID)
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", resp.StatusCode)
		}
	})
}

func getItem(t *testing.T, baseURL string, id uuid.UUID) *http.Response {
	t.Helper()

	urlString := fmt.Sprintf("%s/item/%s", baseURL, id.String())
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, parsedURL.String(), http.NoBody)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	return resp
}
