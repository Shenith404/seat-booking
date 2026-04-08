package email

import (
	"context"

	"github.com/hibiken/asynq"
)

type BookingPayload struct {
	UserEmail  string
	UserName   string
	BookingId  string
	TotalSeats int
	Subject    string
	HtmlBody   string
}

type EmailSendWorker interface {
	SendEmail(ctx context.Context, t *asynq.Task) error
}
