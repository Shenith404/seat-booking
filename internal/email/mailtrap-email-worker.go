package email

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"gopkg.in/gomail.v2"

	"github.com/shenith404/seat-booking/internal/config"
)

type MailTrapEmailWorker struct {
	host       string
	port       int
	user       string
	pass       string
	senderName string
	senderAddr string
}

func NewMailTrapEmailWorker(cfg config.MailTrapConfig) *MailTrapEmailWorker {
	return &MailTrapEmailWorker{
		host:       cfg.Host,
		port:       cfg.Port,
		user:       cfg.User,
		pass:       cfg.Pass,
		senderName: cfg.SenderName,
		senderAddr: cfg.SenderAddr,
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

	m.SetHeader("From", m.FormatAddress(w.senderAddr, w.senderName))
	m.SetHeader("To", payload.UserEmail)
	m.SetHeader("Subject", payload.Subject)
	m.SetBody("text/html", payload.HtmlBody)

	d := gomail.NewDialer(w.host, w.port, w.user, w.pass)
	return d.DialAndSend(m)
}
