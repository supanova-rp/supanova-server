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
	"github.com/google/go-cmp/cmp/cmpopts"
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
				{
					Video: &handlers.AddVideoSectionParams{
						Title:      "Video Section",
						StorageKey: uuid.New().String(),
						Position:   0,
						Type:       domain.SectionTypeVideo,
					},
				},
				{
					Quiz: &handlers.AddQuizSectionParams{
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
					},
				},
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

func TestCourses(t *testing.T) {
	t.Run("courses - returns courses with sections and materials", func(t *testing.T) {
		createdA := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             "Course A",
			Description:       courseDescription,
			CompletionTitle:   courseCompletionTitle,
			CompletionMessage: courseCompletionMessage,
			Materials: []handlers.AddMaterialParams{
				{ID: uuid.New().String(), Name: "Study Guide", StorageKey: uuid.New().String(), Position: 0},
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
							},
						},
					},
				}},
			},
		})

		createdB := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             "Course B",
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

		courses := getCourses(t, testResources.AppURL)

		findCourse := func(id uuid.UUID) *domain.Course {
			for _, c := range courses {
				if c.ID == id {
					return c
				}
			}
			return nil
		}

		foundA := findCourse(createdA.ID)
		if foundA == nil {
			t.Fatalf("course A %s not found in courses response", createdA.ID)
		}

		foundB := findCourse(createdB.ID)
		if foundB == nil {
			t.Fatalf("course B %s not found in courses response", createdB.ID)
		}

		expectedA := &domain.Course{
			ID:                createdA.ID,
			Title:             createdA.Title,
			Description:       createdA.Description,
			CompletionTitle:   createdA.CompletionTitle,
			CompletionMessage: createdA.CompletionMessage,
			Sections: []domain.CourseSection{
				createdA.Sections[0],
				&domain.QuizSection{
					ID:       createdA.Sections[1].GetID(),
					Position: createdA.Sections[1].GetPosition(),
					Type:     domain.SectionTypeQuiz,
				},
			},
			Materials: createdA.Materials,
		}

		expectedB := &domain.Course{
			ID:                createdB.ID,
			Title:             createdB.Title,
			Description:       createdB.Description,
			CompletionTitle:   createdB.CompletionTitle,
			CompletionMessage: createdB.CompletionMessage,
			Sections:          createdB.Sections,
			Materials:         []domain.CourseMaterial{},
		}

		if diff := cmp.Diff(expectedA, foundA); diff != "" {
			t.Errorf("course A mismatch (-want +got):\n%s", diff)
		}

		if diff := cmp.Diff(expectedB, foundB); diff != "" {
			t.Errorf("course B mismatch (-want +got):\n%s", diff)
		}

		deleteCourse(t, testResources.AppURL, createdA.ID)
		deleteCourse(t, testResources.AppURL, createdB.ID)
	})
}

func TestQuiz(t *testing.T) {
	t.Run("quiz questions - happy path", func(t *testing.T) {
		created := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             courseTitle,
			Description:       courseDescription,
			CompletionTitle:   courseCompletionTitle,
			CompletionMessage: courseCompletionMessage,
			Sections: []handlers.AddSectionParams{
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
				},
				},
				{Quiz: &handlers.AddQuizSectionParams{
					Position: 2,
					Type:     domain.SectionTypeQuiz,
					Questions: []handlers.AddQuizQuestionParams{
						{
							Question: "Is this the correct answer?",
							Position: 0,
							Answers: []handlers.AddQuizAnswerParams{
								{Answer: "Yes", IsCorrectAnswer: true, Position: 0},
								{Answer: "No", IsCorrectAnswer: false, Position: 1},
								{Answer: "Maybe", IsCorrectAnswer: false, Position: 2},
							},
						},
						{
							Question:      "Who did it?",
							Position:      1,
							IsMultiAnswer: true,
							Answers: []handlers.AddQuizAnswerParams{
								{Answer: "Me", IsCorrectAnswer: true, Position: 0},
								{Answer: "You", IsCorrectAnswer: false, Position: 1},
								{Answer: "No one", IsCorrectAnswer: true, Position: 2},
							},
						},
					},
				}},
			},
		})

		enrolUserInCourse(t, testResources.AppURL, created.ID)

		expected := []domain.QuizQuestionLegacy{
			{
				Question:      "What is the correct answer?",
				Position:      0,
				IsMultiAnswer: false,
				Answers: []domain.QuizAnswer{
					{Answer: "Correct", IsCorrectAnswer: true, Position: 0},
					{Answer: "Wrong", IsCorrectAnswer: false, Position: 1},
				},
			},
			{
				Question:      "Is this the correct answer?",
				Position:      0,
				IsMultiAnswer: false,
				Answers: []domain.QuizAnswer{
					{Answer: "Yes", IsCorrectAnswer: true, Position: 0},
					{Answer: "No", IsCorrectAnswer: false, Position: 1},
					{Answer: "Maybe", IsCorrectAnswer: false, Position: 2},
				},
			},
			{
				Question:      "Who did it?",
				Position:      1,
				IsMultiAnswer: true,
				Answers: []domain.QuizAnswer{
					{Answer: "Me", IsCorrectAnswer: true, Position: 0},
					{Answer: "You", IsCorrectAnswer: false, Position: 1},
					{Answer: "No one", IsCorrectAnswer: true, Position: 2},
				},
			},
		}

		quizSectionIDs := []uuid.UUID{}
		for _, section := range created.Sections {
			quizSectionIDs = append(quizSectionIDs, section.GetID())
		}
		actual := getQuizQuestions(t, testResources.AppURL, quizSectionIDs)

		// ignore ID, QuizSectionID and Answer->ID
		if diff := cmp.Diff(expected, *actual,
			cmpopts.IgnoreFields(domain.QuizQuestionLegacy{}, "ID", "QuizSectionID"),
			cmpopts.IgnoreFields(domain.QuizAnswer{}, "ID"),
		); diff != "" {
			t.Errorf("quiz questions mismatch (-want +got):\n%s", diff)
		}
	})
}
