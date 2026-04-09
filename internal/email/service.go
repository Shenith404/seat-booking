package email

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type EmailService interface {
	SendBookingEmail(email BookingPayload) error
}

type EmailServiceImpl struct {
	client *asynq.Client
}

const TypeEmailBooking = "email:BookingConfirmation"

func NewService(client *asynq.Client) *EmailServiceImpl {
	return &EmailServiceImpl{
		client: client,
	}
}

func (s *EmailServiceImpl) SendBookingEmail(request BookingPayload) error {

	subject := fmt.Sprintf("Booking Confirmed - Thank you, %s", request.UserName)
	htmlBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8" />
			<meta name="viewport" content="width=device-width, initial-scale=1.0" />
			<title>Booking Confirmation</title>
		</head>
		<body style="margin:0;padding:0;background-color:#f4f6f8;font-family:Arial,sans-serif;color:#1f2937;">
			<table role="presentation" width="100%%" cellspacing="0" cellpadding="0" style="padding:24px 0;background-color:#f4f6f8;">
				<tr>
					<td align="center">
						<table role="presentation" width="600" cellspacing="0" cellpadding="0" style="max-width:600px;width:100%%;background:#ffffff;border-radius:10px;overflow:hidden;box-shadow:0 4px 18px rgba(0,0,0,0.08);">
							<tr>
								<td style="background:#0f172a;padding:20px 24px;color:#ffffff;font-size:20px;font-weight:bold;">
									Sarislabs Seat Booking
								</td>
							</tr>
							<tr>
								<td style="padding:24px;">
									<h2 style="margin:0 0 12px 0;font-size:22px;color:#111827;">Hi %s, your booking is confirmed.</h2>
									<p style="margin:0 0 12px 0;line-height:1.6;color:#374151;">
										Thank you for booking with us. Your seats have been successfully reserved.
									</p>
									<p style="margin:0 0 12px 0;line-height:1.6;color:#374151;">
										Please keep this email for your records. We recommend arriving early to avoid last-minute delays.
									</p>
									<div style="margin-top:20px;padding:14px 16px;background:#f9fafb;border:1px solid #e5e7eb;border-radius:8px;color:#4b5563;font-size:14px;line-height:1.5;">
										Need help? Reply to this email and our support team will assist you.
									</div>
								</td>
							</tr>
							<tr>
								<td style="padding:16px 24px;background:#f3f4f6;color:#6b7280;font-size:12px;line-height:1.5;">
									This is an automated confirmation message from Sarislabs Seat Booking.
								</td>
							</tr>
						</table>
					</td>
				</tr>
			</table>
		</body>
		</html>
	`, request.UserName)

	request.HtmlBody = htmlBody
	request.Subject = subject

	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %w", err)
	}
	task := asynq.NewTask(TypeEmailBooking, payloadBytes)
	// ProcessIn(2 * time.Second) gives a slight delay, useful for deliverability
	_, err = s.client.Enqueue(task)

	if err != nil {
		return fmt.Errorf("failed to enqueue welcome email task: %w", err)
	}

	return nil
}
