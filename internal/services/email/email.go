package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

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
	privateKey string
}

type EmailCourseCompletionParams struct {
	TemplateParams *CourseCompletionParams `json:"template_params"`
	ServiceID      string                  `json:"service_id"`
	TemplateID     string                  `json:"template_id"`
	PublicKey      string                  `json:"user_id"`
	PrivateKey     string                  `json:"accessToken"`
}

func New(cfg *config.EmailService) *Service {
	return &Service{
		serviceID:  cfg.ServiceID,
		templateID: cfg.TemplateID,
		publicKey:  cfg.PublicKey,
		privateKey: cfg.PrivateKey,
	}
}

func (c *Service) SendCourseCompletionNotification(ctx context.Context, params *CourseCompletionParams) error {
	reqBody := &EmailCourseCompletionParams{
		TemplateParams: params,
		ServiceID:      c.serviceID,
		TemplateID:     c.templateID,
		PublicKey:      c.publicKey,
		PrivateKey:     c.privateKey,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	parsedURL, err := url.Parse(emailJSURL)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsedURL.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(res.Body)
		slog.Error(fmt.Sprintf("email service failed with status %d:", res.StatusCode), slog.Any("error", string(bodyBytes)))
	}
	defer res.Body.Close() //nolint:errcheck

	return nil
}
