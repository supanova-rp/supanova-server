package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/supanova-rp/supanova-server/internal/handlers"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
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
		id := insertCourse(ctx, t, testResources)

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

		if result.Title.String != CourseTitle {
			t.Errorf("expected title '%s', got '%s'", CourseTitle, result.Title.String)
		}

		if result.Description.String != CourseDescription {
			t.Errorf("expected description '%s', got '%s'", CourseDescription, result.Description.String)
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
			Title:       CourseTitle,
			Description: CourseDescription,
		}

		resp := addCourse(t, testResources.AppURL, newCourse)
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status 201, got %d", resp.StatusCode)
		}
	})
}
