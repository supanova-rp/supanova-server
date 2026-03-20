//go:build e2e

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

	courseTitle             = "Course A"
	courseDescription       = "Course description"
	courseCompletionTitle   = "Course Complete"
	courseCompletionMessage = "Well done on completing the course"
)

func getUsersAndAssignedCourses(t *testing.T, baseURL string) []domain.UserWithAssignedCourses {
	t.Helper()
	return *postAndParse[[]domain.UserWithAssignedCourses](t, baseURL, "users-to-courses", nil, http.StatusOK)
}

// TODO: Remove once edit course dashboard reuses /courses/overview endpoint
func getCourses(t *testing.T, baseURL string) []*domain.AllCourseLegacy {
	t.Helper()
	return *postAndParse[[]*domain.AllCourseLegacy](t, baseURL, "courses", nil, http.StatusOK)
}

func getCourse(t *testing.T, baseURL string, id uuid.UUID) *domain.Course {
	t.Helper()
	return postAndParse[domain.Course](t, baseURL, "course", map[string]uuid.UUID{"courseId": id}, http.StatusOK)
}

// TODO: Remove once edit course dashboard reuses /courses/overview endpoint
func getQuizQuestions(t *testing.T, baseURL string, quizSectionIDs []uuid.UUID) *[]domain.QuizQuestionLegacy {
	t.Helper()
	return postAndParse[[]domain.QuizQuestionLegacy](t, baseURL, "quiz-questions", map[string][]uuid.UUID{"quizSectionIds": quizSectionIDs}, http.StatusOK)
}

func deleteCourse(t *testing.T, baseURL string, id uuid.UUID) {
	t.Helper()
	postOnly(t, baseURL, "delete-course", map[string]string{"course_id": id.String()}, http.StatusOK)
}

func addCourse(t *testing.T, baseURL string, params *handlers.AddCourseParams) *domain.Course {
	t.Helper()
	return postAndParse[domain.Course](t, baseURL, "add-course", params, http.StatusCreated)
}

func getProgress(t *testing.T, baseURL string, courseID uuid.UUID) *domain.Progress {
	t.Helper()
	return postAndParse[domain.Progress](t, baseURL, "get-progress", map[string]uuid.UUID{"courseId": courseID}, http.StatusOK)
}

func updateProgress(t *testing.T, baseURL string, courseID, sectionID uuid.UUID) {
	t.Helper()
	postOnly(t, baseURL, "update-progress", &handlers.UpdateProgressParams{CourseID: courseID.String(), SectionID: sectionID.String()}, http.StatusNoContent)
}

func getMaterials(t *testing.T, baseURL string, courseID uuid.UUID) []domain.CourseMaterialWithURL {
	t.Helper()
	return *postAndParse[[]domain.CourseMaterialWithURL](t, baseURL, "materials", map[string]string{"courseId": courseID.String()}, http.StatusOK)
}

func resetProgress(t *testing.T, baseURL string, courseID uuid.UUID) {
	t.Helper()
	postOnly(t, baseURL, "reset-progress", &handlers.ResetProgressParams{CourseID: courseID.String()}, http.StatusNoContent)
}

func enrolUserInCourse(t *testing.T, baseURL string, courseID uuid.UUID) {
	t.Helper()
	postOnly(t, baseURL, "update-users-to-courses", &handlers.UpdateCourseEnrolmentParams{UserID: TestUserID, CourseID: courseID.String(), IsEnrolled: false}, http.StatusNoContent)
}

func postAndParse[T any](t *testing.T, baseURL, endpoint string, body any, expectedStatus int) *T {
	t.Helper()
	resp := makePOSTRequest(t, baseURL, endpoint, body)
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
	return parseJSONResponse[T](t, resp)
}

func postOnly(t *testing.T, baseURL, endpoint string, body any, expectedStatus int) {
	t.Helper()
	resp := makePOSTRequest(t, baseURL, endpoint, body)
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != expectedStatus {
		t.Fatalf("expected status %d, got %d", expectedStatus, resp.StatusCode)
	}
}

func register(t *testing.T, baseURL string, params *handlers.RegisterParams) map[string]string {
	t.Helper()

	resp := makePOSTRequest(t, baseURL, "register", params)
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("register failed, expected status 200, got %d", resp.StatusCode)
	}

	return *parseJSONResponse[map[string]string](t, resp)
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
