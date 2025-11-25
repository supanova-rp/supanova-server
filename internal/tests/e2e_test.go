package tests

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	_ "github.com/lib/pq"

	"github.com/supanova-rp/supanova-server/internal/domain"
	"github.com/supanova-rp/supanova-server/internal/handlers"
)

var testResources *TestResources

func TestMain(m *testing.M) {
	ctx := context.Background()

	var err error
	testResources, err = setupTestResources(ctx, &testing.T{})
	if err != nil {
		fmt.Printf("setup tests failed: %s\n", err)
		if testResources != nil {
			testResources.Cleanup(ctx, &testing.T{})
		}
		os.Exit(1)
	}

	exitCode := m.Run()

	testResources.Cleanup(ctx, &testing.T{})
	if exitCode != 0 {
		slog.Error("tests failed", slog.Int("exit_code", exitCode))
	}

	os.Exit(exitCode)
}

func TestCourse(t *testing.T) {
	t.Run("course - happy path", func(t *testing.T) {
		// Add a course
		created := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:       CourseTitle,
			Description: CourseDescription,
		})

		// Retrieve that course
		actual := getCourse(t, testResources.AppURL, created.ID)

		if diff := cmp.Diff(created, actual); diff != "" {
			t.Errorf("course mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("course - not found", func(t *testing.T) {
		nonExistentID := uuid.New()

		resp := makePOSTRequest(t, testResources.AppURL, "course", &handlers.GetCourseParams{
			ID: nonExistentID.String(),
		})
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404, got %d", resp.StatusCode)
		}
	})
}

func TestProgress(t *testing.T) {
	t.Run("user progress - happy path", func(t *testing.T) {
		// TODO: remove this once test is implemented
		t.Skip("Skipping this test until update-progress is implemented")

		_ = os.Getenv("TEST_ENVIRONMENT_USER_ID")
		courseID := uuid.New()

		// TODO: call add-course endpoint

		expected := domain.Progress{
			CompletedSectionIDs: []uuid.UUID{
				uuid.New(),
				uuid.New(),
			},
			CompletedIntro: true,
		}

		// TODO: use update-progress route to insert data

		actual := getProgress(t, testResources.AppURL, courseID)

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}
	})
}
