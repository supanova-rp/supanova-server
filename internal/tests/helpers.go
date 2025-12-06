package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
)

const (
	TestUserID    = "test-user-id"
	testUserName  = "Test User"
	testUserEmail = "test@gmail.com"

	courseTitle       = "Course A"
	courseDescription = "Course description"
)

func getCourse(t *testing.T, baseURL string, id uuid.UUID) *domain.Course {
	t.Helper()

	resp := makePOSTRequest(t, baseURL, "course", map[string]uuid.UUID{
		"courseId": id,
	})
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	return parseJSONResponse[domain.Course](t, resp)
}

func addCourse(t *testing.T, baseURL string, params *handlers.AddCourseParams) *domain.Course {
	t.Helper()

	resp := makePOSTRequest(t, baseURL, "add-course", params)
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", resp.StatusCode)
	}

	return parseJSONResponse[domain.Course](t, resp)
}

func getProgress(t *testing.T, baseURL string, courseID uuid.UUID) *domain.Progress {
	t.Helper()

	resp := makePOSTRequest(t, baseURL, "get-progress", map[string]uuid.UUID{
		"courseId": courseID,
	})
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	return parseJSONResponse[domain.Progress](t, resp)
}

func enrollUserInCourse(t *testing.T, baseURL string, courseID uuid.UUID) {
	t.Helper()

	resp := makePOSTRequest(t, baseURL, "update-users-to-courses", &handlers.UpdateUserCourseEnrollmentParams{
		CourseID:   courseID.String(),
		IsEnrolled: false,
	})
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("enroll failed, expected status 200, got %d", resp.StatusCode)
	}
}

func makePOSTRequest(t *testing.T, baseURL, endpoint string, resource any) *http.Response {
	t.Helper()

	parsedURL, err := url.Parse(fmt.Sprintf("%s/%s/%s", baseURL, config.APIVersion, endpoint))
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
	req.Header.Set("X-Test-User-ID", TestUserID)
	req.Header.Set("X-Test-User-Role", string(config.AdminRole))

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}

	return res
}

func parseJSONResponse[T any](t *testing.T, resp *http.Response) *T {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var result T
	err = json.Unmarshal(body, &result)
	if err != nil {
		t.Fatalf("failed to parse JSON response: %v. Body: %s", err, string(body))
	}

	return &result
}
