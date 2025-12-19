package jobs

import (
	"context"
	"log/slog"

	"github.com/supanova-rp/supanova-server/internal/store"
)

func RetrySend(store *store.Store) func(ctx context.Context) {
	return func(ctx context.Context) {
		// TODO: implement

		failedEmails, err := store.GetEmailFailures(ctx)
		if err != nil {
			slog.Error("failed to fetch failed emails", slog.Any("err", err))
		}

		// TODO: add email_name to migration and schema
		// run sqlc again

		// var emailParams []email.EmailParams
		// loop through failedEmails and transform them into emailParams
		// - based on the failedEmail.emailName we can create 'concrete type', e.g:
		//   - if emailName == "CourseCompletion"
		//         params := CourseCompletionParams
		//         TODO: unmarshall
		//         emailParams = append(emailParams, params)

		// - loop through emails and call Send()
		//     - if success: delete row in email_failures table & log success
		//     - if error:
		//         - log errors for each failure
		//         - if remaining_retries == 1: special log to say that email will be deleted from the email_failures table
		//         - decrement remaining_retries & delete where retries <= 0

	}
}
