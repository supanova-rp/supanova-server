//go:build e2e

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
	testResources, err = setupTestResources(ctx)
	if err != nil {
		fmt.Printf("setup tests failed: %s\n", err)
		if testResources != nil {
			testResources.Cleanup(ctx)
		}
		os.Exit(1)
	}

	exitCode := m.Run()

	testResources.Cleanup(ctx)
	if exitCode != 0 {
		slog.Error("tests failed", slog.Int("exit_code", exitCode))
	}

	os.Exit(exitCode)
}

func TestCourse(t *testing.T) {
	t.Run("course - happy path", func(t *testing.T) {
		created := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             courseTitle,
			Description:       courseDescription,
			CompletionTitle:   courseCompletionTitle,
			CompletionMessage: courseCompletionMessage,
			Materials: []handlers.AddMaterialParams{
				{
					ID:         uuid.New().String(),
					Name:       "Study Guide",
					StorageKey: uuid.New().String(),
					Position:   0,
				},
			},
			Sections: []handlers.AddSectionParams{
				{Video: &handlers.AddVideoSectionParams{
					Title:      "Video Section",
					StorageKey: uuid.New().String(),
					Position:   0,
				}},
				{Quiz: &handlers.AddQuizSectionParams{
					Position: 1,
					Questions: []handlers.AddQuizQuestionParams{
						{
							Question: "What is the correct answer?",
							Position: 0,
							Answers: []handlers.AddQuizAnswerParams{
								{Answer: "Correct", IsCorrectAnswer: true, Position: 0},
								{Answer: "Wrong", IsCorrectAnswer: false, Position: 1},
							},
						},
					},
				}},
			},
		})

		enrolUserInCourse(t, testResources.AppURL, created.ID)

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
		created := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             courseTitle,
			Description:       courseDescription,
			CompletionTitle:   courseCompletionTitle,
			CompletionMessage: courseCompletionMessage,
			Sections: []handlers.AddSectionParams{
				{Video: &handlers.AddVideoSectionParams{
					Title:      "Video Section",
					StorageKey: uuid.New().String(),
					Position:   0,
				}},
			},
		})

		enrolUserInCourse(t, testResources.AppURL, created.ID)

		sectionID := created.Sections[0].GetID()
		updateProgress(t, testResources.AppURL, created.ID, sectionID)

		expected := &domain.Progress{
			CompletedSectionIDs: []uuid.UUID{sectionID},
			CompletedIntro:      false,
		}

		actual := getProgress(t, testResources.AppURL, created.ID)

		if diff := cmp.Diff(expected, actual); diff != "" {
			t.Errorf("progress mismatch (-want +got):\n%s", diff)
		}
	})
}
