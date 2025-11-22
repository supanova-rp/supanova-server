package tests

import (
	"bytes"
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

	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

const (
	courseTitle       = "Course A"
	courseDescription = "Course description"
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

	t.Run("returns course by id", func(t *testing.T) {
		id := uuid.New()
		expectedTitle := courseTitle
		expectedDescription := courseDescription

		_, err := testResources.DB.ExecContext(
			ctx,
			`INSERT INTO courses VALUES ($1, $2, $3)`,
			id,
			expectedTitle,
			expectedDescription,
		)
		if err != nil {
			t.Fatalf("failed to insert test data: %v", err)
		}

		resp := getCourse(t, testResources.AppURL, id)
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		var result sqlc.Course
		err = json.Unmarshal(body, &result)
		if err != nil {
			t.Fatalf("failed to parse JSON response: %v. Body: %s", err, string(body))
		}

		if result.Title.String != expectedTitle {
			t.Errorf("expected title '%s', got '%s'", expectedTitle, result.Title.String)
		}

		if result.Description.String != expectedDescription {
			t.Errorf("expected description '%s', got '%s'", expectedDescription, result.Description.String)
		}

		if result.ID.String() != id.String() {
			t.Errorf("expected id '%s', got '%s'", id.String(), result.ID)
		}
	})

	t.Run("returns not found error", func(t *testing.T) {
		nonExistentID := uuid.New()

		resp := getCourse(t, testResources.AppURL, nonExistentID)
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("creates new course", func(t *testing.T) {
		newCourse := &handlers.AddCourseParams{
			Title:       courseTitle,
			Description: courseDescription,
		}

		resp := addCourse(t, testResources.AppURL, newCourse)
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", resp.StatusCode)
		}
	})
}

func getCourse(t *testing.T, baseURL string, id uuid.UUID) *http.Response {
	t.Helper()

	urlString := fmt.Sprintf("%s/v2/course/%s", baseURL, id.String())
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		t.Fatalf("failed to parse GET /course URL: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, parsedURL.String(), http.NoBody)
	if err != nil {
		t.Fatalf("failed to create GET /course request: %v", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make GET /course request: %v", err)
	}

	return resp
}

func addCourse(t *testing.T, baseURL string, course *handlers.AddCourseParams) *http.Response {
	t.Helper()

	urlString := fmt.Sprintf("%s/v2/course", baseURL)
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		t.Fatalf("failed to parse POST /course URL: %v", err)
	}

	b, err := json.Marshal(course)
	if err != nil {
		t.Fatalf("failed to parse POST /course request body: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, parsedURL.String(), bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("failed to create POST /course request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make POST /course request: %v", err)
	}

	return resp
}
