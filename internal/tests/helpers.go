package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/supanova-rp/supanova-server/internal/handlers"
)

func getCourse(t *testing.T, baseURL string, id uuid.UUID) *http.Response {
	t.Helper()

	return makePOSTRequest(t, baseURL, "course", map[string]uuid.UUID{
		"id": id,
	})
}

func addCourse(t *testing.T, baseURL string, course *handlers.AddCourseParams) *http.Response {
	t.Helper()

	return makePOSTRequest(t, baseURL, "add-course", course)
}

func makePOSTRequest(t *testing.T, baseURL, endpoint string, resource any) *http.Response {
	t.Helper()

	parsedURL, err := url.Parse(fmt.Sprintf("%s/v2/%s", baseURL, endpoint))
	if err != nil {
		t.Fatalf("failed to parse URL: %v", err)
	}

	b, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}

	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, parsedURL.String(), bytes.NewBuffer(b))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	return res
}

func insertCourse(ctx context.Context, t *testing.T, testResources *TestResources) uuid.UUID {
	t.Helper()
	id := uuid.New()

	_, err := testResources.DB.ExecContext(
		ctx,
		`INSERT INTO courses VALUES ($1, $2, $3)`,
		id,
		CourseTitle,
		CourseDescription,
	)
	if err != nil {
		t.Fatalf("failed to insert test data: %v", err)
	}

	return id
}
