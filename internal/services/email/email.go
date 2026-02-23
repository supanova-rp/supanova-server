package email

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/mailgun/mailgun-go/v5"

	"github.com/supanova-rp/supanova-server/internal/config"
	"github.com/supanova-rp/supanova-server/internal/handlers/errors"
	"github.com/supanova-rp/supanova-server/internal/services/cron"
	"github.com/supanova-rp/supanova-server/internal/store/sqlc"
	"github.com/supanova-rp/supanova-server/internal/utils"
)

type EmailService struct {
	client        *mailgun.Client
	sender        string
	recipient     string
	domain        string
	templateNames *TemplateNames
	emailNames    *EmailNames
	store         EmailRepository
	retryCron     *cron.Cron
	stopRetry     context.CancelFunc
}

type EmailRepository interface {
	AddFailedEmail(context.Context, sqlc.AddFailedEmailParams) error
	GetFailedEmails(context.Context) ([]FailedEmail, error)
	UpdateFailedEmail(context.Context, sqlc.UpdateFailedEmailParams) error
	DeleteFailedEmail(context.Context, pgtype.UUID) error
}

type FailedEmail struct {
	ID             uuid.UUID
	TemplateParams []byte
	TemplateName   string
	EmailName      string
	Retries        int
}

type EmailParams interface {
	ToTemplateVariables() map[string]string
}

type TemplateNames struct {
	CourseCompletion string
}

type EmailNames struct {
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

func New(cfg *config.EmailService, store EmailRepository) (*EmailService, error) {
	mg := mailgun.NewMailgun(cfg.SendingKey)
	// TODO: set EU domain once we have a Supanova email domain
	// err := mg.SetAPIBase(mailgun.APIBaseEU)

	retryCron := cron.New(cfg.CronSchedule, "email-retry")

	service := &EmailService{
		client:    mg,
		domain:    cfg.Domain,
		sender:    cfg.Sender,
		recipient: cfg.Recipient,
		templateNames: &TemplateNames{
			CourseCompletion: cfg.CourseCompletionTemplateName,
		},
		emailNames: &EmailNames{
			CourseCompletion: "course-completion",
		},
		store:     store,
		retryCron: retryCron,
	}

	stopRetry, err := service.SetupRetry()
	if err != nil {
		return nil, err
	}
	service.stopRetry = stopRetry

	return service, nil
}

func (e *EmailService) GetTemplateNames() *TemplateNames {
	return e.templateNames
}

func (e *EmailService) GetEmailNames() *EmailNames {
	return e.emailNames
}

func (e *EmailService) SetupRetry() (context.CancelFunc, error) {
	return e.retryCron.Setup(e.RetryJob())
}

func (e *EmailService) AddFailedEmail(ctx context.Context, err error, templateParams EmailParams, templateName, emailName string) {
	paramBytes, marshalErr := json.Marshal(templateParams)
	if marshalErr != nil {
		slog.Error("failed to marshal template params", slog.Any("error", marshalErr))
		return
	}

	sqlcParams := sqlc.AddFailedEmailParams{
		Error:          err.Error(),
		TemplateName:   templateName,
		TemplateParams: paramBytes,
		EmailName:      emailName,
	}

	err = e.store.AddFailedEmail(ctx, sqlcParams)
	if err != nil {
		slog.Error("failed to add email to DB", slog.Any("error", err))
	} else {
		slog.Debug("added email to DB")
	}
}

func (e *EmailService) Send(ctx context.Context, params EmailParams, templateName, emailName string) (err error) {
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
			e.AddFailedEmail(ctx, err, params, templateName, emailName)
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

type RetryParams struct {
	ID             pgtype.UUID
	templateParams EmailParams
	templateName   string
	emailName      string
	retries        int
}

func (e *EmailService) RetryJob() func(ctx context.Context) {
	return func(ctx context.Context) {
		failedEmails, err := e.store.GetFailedEmails(ctx)

		if err != nil {
			slog.Error(errors.Getting("failed emails"), slog.Any("error", err))
			return
		}

		if len(failedEmails) == 0 {
			slog.Debug("no failed emails to retry")
			return
		}

		var sendParams []RetryParams

		for _, fe := range failedEmails {
			switch fe.EmailName {
			case e.GetEmailNames().CourseCompletion:
				sendParams = appendParams[*CourseCompletionParams](
					&fe,
					sendParams,
				)
			default:
				slog.Error("email name not found", slog.String("email_name", fe.EmailName))
			}
		}

		for _, sendP := range sendParams {
			err := e.RetrySend(ctx, &sendP)
			if err != nil {
				slog.Error("email retry failed", slog.Any("error", err))
				e.handleRetryFailure(ctx, &sendP, err)
			} else {
				slog.Debug("email retry success", slog.String("id", sendP.ID.String()))
				e.deleteFailedEmail(ctx, sendP.ID)
			}
		}
	}
}

func appendParams[T EmailParams](params *FailedEmail, sendParams []RetryParams) []RetryParams {
	var templateParams T
	if err := json.Unmarshal(params.TemplateParams, &templateParams); err != nil {
		slog.Error("failed to unmarshal template params", slog.Any("error", err))
		return sendParams
	}

	pgUUID, err := utils.PGUUIDFrom(params.ID.String())
	if err != nil {
		slog.Error("failed to parse failed email id", slog.Any("error", err), slog.String("id", params.ID.String()))
		return sendParams
	}

	sendParams = append(sendParams, RetryParams{
		ID:             pgUUID,
		templateName:   params.TemplateName,
		templateParams: templateParams,
		emailName:      params.EmailName,
		retries:        params.Retries,
	})

	return sendParams
}

func (e *EmailService) RetrySend(ctx context.Context, params *RetryParams) error {
	message := mailgun.NewMessage(
		e.domain,
		e.sender,
		"", // subject set by template
		"", // text set by template,
		e.recipient,
	)

	message.SetTemplate(params.templateName)

	for key, value := range params.templateParams.ToTemplateVariables() {
		if err := message.AddTemplateVariable(key, value); err != nil {
			return err
		}
	}

	_, err := e.client.Send(ctx, message)
	if err != nil {
		return err
	}

	return nil
}

func (e *EmailService) deleteFailedEmail(ctx context.Context, id pgtype.UUID) {
	err := e.store.DeleteFailedEmail(ctx, id)
	if err != nil {
		slog.Error("failed to delete email", slog.Any("error", err))
		return
	}

	slog.Debug("delete email success", slog.String("id", id.String()))
}

func (e *EmailService) handleRetryFailure(ctx context.Context, retryParams *RetryParams, err error) {
	if retryParams.retries > 0 {
		e.updateFailedEmail(ctx, retryParams, err)
		return
	}

	slog.Debug("no retries left")
	e.deleteFailedEmail(ctx, retryParams.ID)
}

func (e *EmailService) updateFailedEmail(ctx context.Context, retryParams *RetryParams, err error) {
	sqlcParams := sqlc.UpdateFailedEmailParams{
		Retries: int32(retryParams.retries) - 1, // #nosec G115 retry value is guaranteed to be small because 0 <= value >= 5
		Error:   err.Error(),
	}

	err = e.store.UpdateFailedEmail(ctx, sqlcParams)
	if err != nil {
		slog.Error("failed to update email", slog.String("id", retryParams.ID.String()), slog.Any("error", err))
		return
	}

	slog.Debug("updated failed email")
}

func (e *EmailService) StopRetry(ctx context.Context) {
	e.stopRetry() // cancel cron contexts to prevent new jobs from starting

	stopRetryCtx := e.retryCron.Stop() // returns a context that waits until existing cron jobs finish
	<-stopRetryCtx.Done()
	slog.Info("email retry cron jobs completed")
}
