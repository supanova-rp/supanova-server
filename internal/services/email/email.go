package email

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mailgun/mailgun-go/v5"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
)

type EmailService struct {
	client        *mailgun.Client
	sender        string
	recipient     string
	domain        string
	templateNames *TemplateNames
	store         EmailRepository
}

type EmailRepository interface {
	AddEmailFailure(context.Context, sqlc.AddEmailFailureParams) error
	GetEmailFailures(context.Context) ([]FailedEmail, error)
	UpdateEmailFailure(context.Context, sqlc.UpdateEmailFailureParams) error
	DeleteEmailFailures(context.Context, []pgtype.UUID) error
}

type EmailParams interface {
	ToTemplateVariables() map[string]string
}

type FailedEmail struct {
	ID             uuid.UUID
	TemplateParams []byte
	TemplateName   string
	EmailName      string
}

type TemplateNames struct {
	CourseCompletion string
}

type CourseCompletionParams struct {
	CourseName          string `json:"course_name"`
	UserName            string `json:"user_name"`
	UserEmail           string `json:"user_email"`
	CompletionTimestamp string `json:"completion_timestamp"`
}

func (p *CourseCompletionParams) ToTemplateVariables() map[string]string {
	return map[string]string{
		"course_name":          p.CourseName,
		"user_name":            p.UserName,
		"user_email":           p.UserEmail,
		"completion_timestamp": p.CompletionTimestamp,
	}
}

func New(cfg *config.EmailService, store EmailRepository) *EmailService {
	mg := mailgun.NewMailgun(cfg.SendingKey)

	// TODO: set EU domain once we have a Supanova email domain
	// err := mg.SetAPIBase(mailgun.APIBaseEU)

	return &EmailService{
		client:    mg,
		domain:    cfg.Domain,
		sender:    cfg.Sender,
		recipient: cfg.Recipient,
		templateNames: &TemplateNames{
			CourseCompletion: cfg.CourseCompletionTemplateName,
		},
		store: store,
	}
}

func (e *EmailService) GetTemplateNames() *TemplateNames {
	return e.templateNames
}

func (e *EmailService) Send(ctx context.Context, params EmailParams, templateName string) (err error) {
	message := mailgun.NewMessage(
		e.domain,
		e.sender,
		"", // subject set by template
		"", // text set by template,
		e.recipient,
	)

	message.SetTemplate(templateName)

	defer func() {
		if err != nil {
			e.AddEmailFailure(ctx, err, params, templateName)
		}
	}()

	for key, value := range params.ToTemplateVariables() {
		if err := message.AddTemplateVariable(key, value); err != nil {
			return err
		}
	}

	_, err = e.client.Send(ctx, message)
	return err
}

func (e *EmailService) AddEmailFailure(ctx context.Context, err error, templateParams EmailParams, templateName string) {
	paramBytes, marshalErr := json.Marshal(templateParams)
	if marshalErr != nil {
		slog.Error("failed to marshal template params", slog.Any("err", marshalErr))
		return
	}

	sqlcParams := sqlc.AddEmailFailureParams{
		Error:          err.Error(),
		TemplateName:   templateName,
		TemplateParams: paramBytes,
	}

	err = e.store.AddEmailFailure(ctx, sqlcParams)
	slog.Error("failed to add failed email to email_failures table", slog.Any("err", err))
}
