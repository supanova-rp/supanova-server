package email

import (
	"context"

	"github.com/mailgun/mailgun-go/v5"

	"github.com/supanova-rp/supanova-server/internal/config"
)

type EmailService struct {
	client       *mailgun.Client
	sender       string
	recipient    string
	domain       string
	templateName string
}

type CourseCompletionParams struct {
	UserName            string
	UserEmail           string
	CourseName          string
	CompletionTimestamp string
}

func New(cfg *config.EmailService) *EmailService {
	mg := mailgun.NewMailgun(cfg.SendingKey)

	// TODO: set EU domain once we have a Supanova email domain
	// err := mg.SetAPIBase(mailgun.APIBaseEU)

	return &EmailService{
		client:       mg,
		domain:       cfg.Domain,
		sender:       cfg.Sender,
		recipient:    cfg.Recipient,
		templateName: cfg.TemplateName,
	}
}

func (s *EmailService) SendCourseCompletionNotification(ctx context.Context, params *CourseCompletionParams) error {
	message := mailgun.NewMessage(
		s.domain,
		s.sender,
		"", // subject set by template
		"", // text set by template,
		s.recipient,
	)

	message.SetTemplate(s.templateName)
	templateParams := map[string]string{
		"course_name":          params.CourseName,
		"user_name":            params.UserName,
		"user_email":           params.UserEmail,
		"completion_timestamp": params.CompletionTimestamp,
	}
	for key, value := range templateParams {
		if err := message.AddTemplateVariable(key, value); err != nil {
			return err
		}
	}

	// TODO: log if email was successful
	_, err := s.client.Send(ctx, message)

	return err
}
