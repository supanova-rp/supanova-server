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
	t.Run("course happy path - add, get, delete", func(t *testing.T) {
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
					Type:       domain.SectionTypeVideo,
				}},
				{Quiz: &handlers.AddQuizSectionParams{
					Position: 1,
					Type:     domain.SectionTypeQuiz,
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

		deleteCourse(t, testResources.AppURL, created.ID)

		resp := makePOSTRequest(t, testResources.AppURL, "course", &handlers.GetCourseParams{
			ID: created.ID.String(),
		})
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404 after deletion, got %d", resp.StatusCode)
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

func TestMaterials(t *testing.T) {
	t.Run("materials happy path", func(t *testing.T) {
		materialID := uuid.New()
		storageKey := uuid.New()

		created := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             courseTitle,
			Description:       courseDescription,
			CompletionTitle:   courseCompletionTitle,
			CompletionMessage: courseCompletionMessage,
			Materials: []handlers.AddMaterialParams{
				{
					ID:         materialID.String(),
					Name:       "Study Guide",
					StorageKey: storageKey.String(),
					Position:   0,
				},
			},
		})

		enrolUserInCourse(t, testResources.AppURL, created.ID)

		materials := getMaterials(t, testResources.AppURL, created.ID)

		expectedURL := fmt.Sprintf("https://cdn.example.com/%s/materials/%s.pdf", created.ID.String(), storageKey.String())
		expected := []domain.CourseMaterialWithURL{
			{
				ID:       materialID,
				Name:     "Study Guide",
				Position: 0,
				URL:      expectedURL,
			},
		}

		if diff := cmp.Diff(expected, materials); diff != "" {
			t.Errorf("materials mismatch (-want +got):\n%s", diff)
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
					Type:       domain.SectionTypeVideo,
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

		resetProgress(t, testResources.AppURL, created.ID)

		afterReset := getProgress(t, testResources.AppURL, created.ID)

		expectedAfterReset := &domain.Progress{
			CompletedSectionIDs: nil,
			CompletedIntro:      false,
		}

		if diff := cmp.Diff(expectedAfterReset, afterReset); diff != "" {
			t.Errorf("progress after reset mismatch (-want +got):\n%s", diff)
		}
	})
}
