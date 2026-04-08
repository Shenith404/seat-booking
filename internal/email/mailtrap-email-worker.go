package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"gopkg.in/gomail.v2"

	// We import email just to use the Type constants and Structs
	"github.com/shenith404/seat-booking/internal/config"
)

type MailTrapEmailWorker struct {
	MailtrapHost string
	MailtrapPort int
	MailtrapUser string
	MailtrapPass string
}

func NewMailTrapEmailWorker(cfg config.MailTrapConfig) *MailTrapEmailWorker {
	return &MailTrapEmailWorker{

		MailtrapHost: cfg.MailtrapHost,
		MailtrapPort: cfg.MailtrapPort,
		MailtrapUser: cfg.MailtrapUser,
		MailtrapPass: cfg.MailtrapPass,
	}
}

// inherit the EmailSendWorker interface
func (w *MailTrapEmailWorker) SendEmail(ctx context.Context, t *asynq.Task) error {
	var payload BookingPayload

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("json.Unmarshal failed: %v: %w", err, asynq.SkipRetry)
	}

	log.Printf("[Worker] Processing booking confirmation email for: %s...\n", payload.UserEmail)

	// Send the email using Gomail
	err := w.sendEmailViaMailtrap(payload)
	if err != nil {
		log.Printf("[Worker] Failed to send email to %s: %v\n", payload.UserEmail, err)
		return err // Returning an error tells Asynq to retry later
	}

	log.Printf("[Worker] Successfully sent email to %s!\n", payload.UserEmail)
	return nil
}

func (w *MailTrapEmailWorker) sendEmailViaMailtrap(payload BookingPayload) error {
	m := gomail.NewMessage()

	// Using your company name for the sender
	m.SetHeader("From", "noreply@sarislabs.com")
	m.SetHeader("To", payload.UserEmail)
	m.SetHeader("Subject", payload.Subject)

	// HTML Body
	htmlBody := payload.HtmlBody
	m.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(w.MailtrapHost, w.MailtrapPort, w.MailtrapUser, w.MailtrapPass)
	return d.DialAndSend(m)
}
