package email

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/supanova-rp/supanova-server/internal/config"
)

const emailJSURL = "https://api.emailjs.com/api/v1.0/email/send"

type CourseCompletionParams struct {
	UserName            string `json:"user_name"`
	UserEmail           string `json:"user_email"`
	CourseName          string `json:"course_name"`
	CompletionTimestamp string `json:"completion_timestamp"`
}

type Service struct {
	serviceID  string
	templateID string
	publicKey  string
}

type EmailCourseCompletionParams struct {
	TemplateParams *CourseCompletionParams `json:"template_params"`
	ServiceID      string                  `json:"service_id"`
	TemplateID     string                  `json:"template_id"`
	PublicKey      string                  `json:"user_id"`
}

func New(cfg *config.EmailService) *Service {
	return &Service{
		serviceID:  cfg.ServiceID,
		templateID: cfg.TemplateID,
		publicKey:  cfg.PublicKey,
	}
}

func (c *Service) SendCourseCompletion(ctx echo.Context, params *CourseCompletionParams) error {
	reqBody := &EmailCourseCompletionParams{
		TemplateParams: params,
		ServiceID:      c.serviceID,
		TemplateID:     c.templateID,
		PublicKey:      c.publicKey,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(emailJSURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
