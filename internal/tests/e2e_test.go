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

func TestAuth(t *testing.T) {
	t.Run("register - happy path", func(t *testing.T) {
		result := register(t, testResources.AppURL, &handlers.RegisterParams{
			Name:     "New User",
			Email:    "newuser@example.com",
			Password: "password123",
		})

		if result["newUserId"] == "" {
			t.Fatal("expected newUserId in response, got empty string")
		}
	})

	t.Run("register - missing required fields", func(t *testing.T) {
		resp := makePOSTRequest(t, testResources.AppURL, "register", &handlers.RegisterParams{
			Name: "New User",
		})
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", resp.StatusCode)
		}
	})
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

		users := getUsersAndAssignedCourses(t, testResources.AppURL)

		var testUser *domain.UserWithAssignedCourses
		for i := range users {
			if users[i].ID == TestUserID {
				testUser = &users[i]
				break
			}
		}

		if testUser == nil {
			t.Fatalf("test user %s not found in users-to-courses response", TestUserID)
		}

		var foundCourseID *uuid.UUID
		for i := range testUser.CourseIDs {
			if testUser.CourseIDs[i] == created.ID {
				foundCourseID = &testUser.CourseIDs[i]
				break
			}
		}

		if foundCourseID == nil {
			t.Fatalf("expected course %s to be in test user's courses, got %v", created.ID, testUser.CourseIDs)
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

		expectedProgress := &domain.Progress{
			CompletedSectionIDs: []uuid.UUID{sectionID},
			CompletedIntro:      false,
		}

		actualProgress := getProgress(t, testResources.AppURL, created.ID)

		if diff := cmp.Diff(expectedProgress, actualProgress); diff != "" {
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

		actualIntroCompleted := setIntroCompleted(t, testResources.AppURL, created.ID)

		expectedIntroCompleted := &handlers.SetIntroCompletedResponse{
			CourseID:       created.ID.String(),
			CompletedIntro: true,
		}

		if diff := cmp.Diff(expectedIntroCompleted, actualIntroCompleted); diff != "" {
			t.Errorf("set intro completed response mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestEditCourse(t *testing.T) {
	t.Run("updates course fields, sections, and materials", func(t *testing.T) {
		videoStorageKey := uuid.New()
		materialID := uuid.New()
		materialStorageKey := uuid.New()

		created := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             "Original Title",
			Description:       "Original Description",
			CompletionTitle:   "Original Completion Title",
			CompletionMessage: "Original Completion Message",
			Materials: []handlers.AddMaterialParams{
				{ID: materialID.String(), Name: "Original Guide", StorageKey: materialStorageKey.String(), Position: 0},
			},
			Sections: []handlers.AddSectionParams{
				{Video: &handlers.AddVideoSectionParams{
					Type:       domain.SectionTypeVideo,
					Title:      "Original Video",
					StorageKey: videoStorageKey.String(),
					Position:   0,
				}},
				{Quiz: &handlers.AddQuizSectionParams{
					Type:     domain.SectionTypeQuiz,
					Position: 1,
					Questions: []handlers.AddQuizQuestionParams{
						{
							Question: "Original question?",
							Position: 0,
							Answers: []handlers.AddQuizAnswerParams{
								{Answer: "Original answer", IsCorrectAnswer: true, Position: 0},
							},
						},
					},
				}},
			},
		})

		existingVideo, ok := created.Sections[0].(*domain.VideoSection)
		if !ok {
			t.Fatal("expected first section to be a VideoSection")
		}
		existingQuiz, ok := created.Sections[1].(*domain.QuizSection)
		if !ok {
			t.Fatal("expected second section to be a QuizSection")
		}
		existingQuestion := existingQuiz.Questions[0]
		existingAnswer := existingQuestion.Answers[0]

		newMaterialID := uuid.New()
		newMaterialStorageKey := uuid.New()
		newVideoStorageKey := uuid.New()

		edited := editCourse(t, testResources.AppURL, &handlers.EditCourseRequest{
			CourseID: created.ID.String(),
			EditedCourse: handlers.EditedCourseFields{
				Title:             "Updated Title",
				Description:       "Updated Description",
				CompletionTitle:   "Updated Completion Title",
				CompletionMessage: "Updated Completion Message",
				Materials: []handlers.EditMaterialParams{
					{ID: materialID.String(), Name: "Updated Guide", StorageKey: materialStorageKey.String(), Position: 0},
					{ID: newMaterialID.String(), Name: "New Guide", StorageKey: newMaterialStorageKey.String(), Position: 1},
				},
				Sections: []handlers.EditSectionParams{
					{Video: &handlers.EditVideoSectionParams{
						Type:         domain.SectionTypeVideo,
						ID:           existingVideo.ID.String(),
						IsNewSection: false,
						Title:        "Updated Video",
						StorageKey:   videoStorageKey.String(),
						Position:     0,
					}},
					{Quiz: &handlers.EditQuizSectionParams{
						Type:         domain.SectionTypeQuiz,
						ID:           existingQuiz.ID.String(),
						IsNewSection: false,
						Position:     1,
						Questions: []handlers.EditQuizQuestionParams{
							{
								ID:       existingQuestion.ID.String(),
								Question: "Updated question?",
								Position: 0,
								Answers: []handlers.EditQuizAnswerParams{
									{ID: existingAnswer.ID.String(), Answer: "Updated answer", IsCorrectAnswer: true, Position: 0},
								},
							},
						},
					}},
					{Video: &handlers.EditVideoSectionParams{
						Type:         domain.SectionTypeVideo,
						IsNewSection: true,
						Title:        "New Video",
						StorageKey:   newVideoStorageKey.String(),
						Position:     2,
					}},
				},
			},
			DeletedSectionIDs: handlers.DeletedSectionIDs{
				VideoSectionIDs: []string{},
				QuizSectionIDs:  []string{},
				QuestionIDs:     []string{},
				AnswerIDs:       []string{},
			},
			DeletedMaterialIDs: []string{},
		})

		if edited.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got %q", edited.Title)
		}
		if edited.Description != "Updated Description" {
			t.Errorf("expected description 'Updated Description', got %q", edited.Description)
		}
		if edited.CompletionTitle != "Updated Completion Title" {
			t.Errorf("expected completionTitle 'Updated Completion Title', got %q", edited.CompletionTitle)
		}
		if edited.CompletionMessage != "Updated Completion Message" {
			t.Errorf("expected completionMessage 'Updated Completion Message', got %q", edited.CompletionMessage)
		}

		if len(edited.Sections) != 3 {
			t.Fatalf("expected 3 sections, got %d", len(edited.Sections))
		}

		updatedVideo, ok := edited.Sections[0].(*domain.VideoSection)
		if !ok {
			t.Fatal("expected first section to be a VideoSection")
		}
		if updatedVideo.ID != existingVideo.ID {
			t.Errorf("expected video section ID to be unchanged, got %s", updatedVideo.ID)
		}
		if updatedVideo.Title != "Updated Video" {
			t.Errorf("expected video title 'Updated Video', got %q", updatedVideo.Title)
		}

		updatedQuiz, ok := edited.Sections[1].(*domain.QuizSection)
		if !ok {
			t.Fatal("expected second section to be a QuizSection")
		}
		if updatedQuiz.ID != existingQuiz.ID {
			t.Errorf("expected quiz section ID to be unchanged, got %s", updatedQuiz.ID)
		}
		if len(updatedQuiz.Questions) != 1 {
			t.Fatalf("expected 1 question, got %d", len(updatedQuiz.Questions))
		}
		if updatedQuiz.Questions[0].Question != "Updated question?" {
			t.Errorf("expected question 'Updated question?', got %q", updatedQuiz.Questions[0].Question)
		}
		if updatedQuiz.Questions[0].Answers[0].Answer != "Updated answer" {
			t.Errorf("expected answer 'Updated answer', got %q", updatedQuiz.Questions[0].Answers[0].Answer)
		}

		newVideo, ok := edited.Sections[2].(*domain.VideoSection)
		if !ok {
			t.Fatal("expected third section to be a VideoSection")
		}
		if newVideo.Title != "New Video" {
			t.Errorf("expected new video title 'New Video', got %q", newVideo.Title)
		}
		if newVideo.StorageKey != newVideoStorageKey {
			t.Errorf("expected new video storage key %s, got %s", newVideoStorageKey, newVideo.StorageKey)
		}

		if len(edited.Materials) != 2 {
			t.Fatalf("expected 2 materials, got %d", len(edited.Materials))
		}

		findMaterial := func(id uuid.UUID) *domain.CourseMaterial {
			for i := range edited.Materials {
				if edited.Materials[i].ID == id {
					return &edited.Materials[i]
				}
			}
			return nil
		}

		updatedMaterial := findMaterial(materialID)
		if updatedMaterial == nil {
			t.Fatalf("expected existing material %s to be present", materialID)
		}
		if updatedMaterial.Name != "Updated Guide" {
			t.Errorf("expected material name 'Updated Guide', got %q", updatedMaterial.Name)
		}

		newMaterial := findMaterial(newMaterialID)
		if newMaterial == nil {
			t.Fatalf("expected new material %s to be present", newMaterialID)
		}
		if newMaterial.Name != "New Guide" {
			t.Errorf("expected new material name 'New Guide', got %q", newMaterial.Name)
		}

		deleteCourse(t, testResources.AppURL, created.ID)
	})

	t.Run("deleted sections are removed from user progress", func(t *testing.T) {
		created := addCourse(t, testResources.AppURL, &handlers.AddCourseParams{
			Title:             courseTitle,
			Description:       courseDescription,
			CompletionTitle:   courseCompletionTitle,
			CompletionMessage: courseCompletionMessage,
			Sections: []handlers.AddSectionParams{
				{Video: &handlers.AddVideoSectionParams{
					Type:       domain.SectionTypeVideo,
					Title:      "Video To Delete",
					StorageKey: uuid.New().String(),
					Position:   0,
				}},
				{Video: &handlers.AddVideoSectionParams{
					Type:       domain.SectionTypeVideo,
					Title:      "Video To Keep",
					StorageKey: uuid.New().String(),
					Position:   1,
				}},
			},
		})

		enrolUserInCourse(t, testResources.AppURL, created.ID)

		sectionToDelete := created.Sections[0].GetID()
		sectionToKeep := created.Sections[1].GetID()

		updateProgress(t, testResources.AppURL, created.ID, sectionToDelete)
		updateProgress(t, testResources.AppURL, created.ID, sectionToKeep)

		videoToKeep, ok := created.Sections[1].(*domain.VideoSection)
		if !ok {
			t.Fatal("expected second section to be a VideoSection")
		}

		editCourse(t, testResources.AppURL, &handlers.EditCourseRequest{
			CourseID: created.ID.String(),
			EditedCourse: handlers.EditedCourseFields{
				Title:             created.Title,
				Description:       created.Description,
				CompletionTitle:   created.CompletionTitle,
				CompletionMessage: created.CompletionMessage,
				Materials:         []handlers.EditMaterialParams{},
				Sections: []handlers.EditSectionParams{
					{Video: &handlers.EditVideoSectionParams{
						Type:         domain.SectionTypeVideo,
						ID:           videoToKeep.ID.String(),
						IsNewSection: false,
						Title:        videoToKeep.Title,
						StorageKey:   videoToKeep.StorageKey.String(),
						Position:     videoToKeep.Position,
					}},
				},
			},
			DeletedSectionIDs: handlers.DeletedSectionIDs{
				VideoSectionIDs: []string{sectionToDelete.String()},
				QuizSectionIDs:  []string{},
				QuestionIDs:     []string{},
				AnswerIDs:       []string{},
			},
			DeletedMaterialIDs: []string{},
		})

		progress := getProgress(t, testResources.AppURL, created.ID)

		for _, id := range progress.CompletedSectionIDs {
			if id == sectionToDelete {
				t.Errorf("expected deleted section %s to be removed from progress", sectionToDelete)
			}
		}

		foundKept := false
		for _, id := range progress.CompletedSectionIDs {
			if id == sectionToKeep {
				foundKept = true
				break
			}
		}
		if !foundKept {
			t.Errorf("expected kept section %s to remain in progress", sectionToKeep)
		}

		deleteCourse(t, testResources.AppURL, created.ID)
	})

	t.Run("missing required fields returns 400", func(t *testing.T) {
		resp := makePOSTRequest(t, testResources.AppURL, "edit-course", map[string]string{
			"edited_course_id": uuid.New().String(),
		})
		defer resp.Body.Close() //nolint:errcheck

		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", resp.StatusCode)
		}
	})
}

// TODO: Remove once edit course dashboard reuses /courses/overview endpoint
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

		findCourse := func(id uuid.UUID) *domain.AllCourseLegacy {
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

		expectedA := &domain.AllCourseLegacy{
			ID:                createdA.ID,
			Title:             createdA.Title,
			Description:       createdA.Description,
			CompletionTitle:   createdA.CompletionTitle,
			CompletionMessage: createdA.CompletionMessage,
			Sections: []domain.CourseSection{
				createdA.Sections[0],
				&domain.QuizSectionLegacy{
					ID:       createdA.Sections[1].GetID(),
					Position: createdA.Sections[1].GetPosition(),
					Type:     domain.SectionTypeQuiz,
				},
			},
			Materials: createdA.Materials,
		}

		expectedB := &domain.AllCourseLegacy{
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

// TODO: Remove once edit course dashboard reuses /courses/overview endpoint
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
