package worker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// BookingJob represents a job to process after booking
type BookingJob struct {
	BookingID     uuid.UUID
	CustomerEmail string
	CustomerPhone string
	ShowID        uuid.UUID
	SeatIDs       []uuid.UUID
	TicketIDs     []uuid.UUID
}

// QRCodeResult represents the result of QR code generation
type QRCodeResult struct {
	TicketID uuid.UUID
	QRHash   string
}

// EmailResult represents the result of email sending
type EmailResult struct {
	BookingID uuid.UUID
	Sent      bool
	Error     error
}

// Worker processes background jobs
type Worker struct {
	jobQueue    chan BookingJob
	workerCount int
	wg          sync.WaitGroup
	qrCallback  func(ctx context.Context, ticketID uuid.UUID, qrHash string) error
}

// NewWorker creates a new background worker
func NewWorker(workerCount int, queueSize int) *Worker {
	return &Worker{
		jobQueue:    make(chan BookingJob, queueSize),
		workerCount: workerCount,
	}
}

// SetQRCallback sets the callback for updating QR hash in database
func (w *Worker) SetQRCallback(cb func(ctx context.Context, ticketID uuid.UUID, qrHash string) error) {
	w.qrCallback = cb
}

// Start starts the worker pool
func (w *Worker) Start(ctx context.Context) {
	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(ctx, i)
	}
	log.Printf("Started %d background workers", w.workerCount)
}

// Stop gracefully stops all workers
func (w *Worker) Stop() {
	close(w.jobQueue)
	w.wg.Wait()
	log.Println("All background workers stopped")
}

// Submit submits a job to the worker pool
func (w *Worker) Submit(job BookingJob) error {
	select {
	case w.jobQueue <- job:
		log.Printf("Job submitted for booking %s", job.BookingID)
		return nil
	default:
		return fmt.Errorf("job queue is full")
	}
}

// worker processes jobs from the queue
func (w *Worker) worker(ctx context.Context, id int) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d shutting down", id)
			return
		case job, ok := <-w.jobQueue:
			if !ok {
				log.Printf("Worker %d: job queue closed", id)
				return
			}
			w.processJob(ctx, job)
		}
	}
}

// processJob processes a single booking job
func (w *Worker) processJob(ctx context.Context, job BookingJob) {
	log.Printf("Processing booking %s", job.BookingID)

	// Generate QR codes for each ticket
	for _, ticketID := range job.TicketIDs {
		qrHash := generateQRHash(job.BookingID, ticketID)

		if w.qrCallback != nil {
			if err := w.qrCallback(ctx, ticketID, qrHash); err != nil {
				log.Printf("Failed to update QR hash for ticket %s: %v", ticketID, err)
			}
		}

		log.Printf("Generated QR hash for ticket %s: %s", ticketID, qrHash[:16]+"...")
	}

	// Send confirmation email (simulated)
	if err := sendConfirmationEmail(job); err != nil {
		log.Printf("Failed to send email for booking %s: %v", job.BookingID, err)
	} else {
		log.Printf("Confirmation email sent for booking %s to %s", job.BookingID, job.CustomerEmail)
	}
}

// generateQRHash generates a unique QR code hash for a ticket
func generateQRHash(bookingID, ticketID uuid.UUID) string {
	data := fmt.Sprintf("%s:%s:%d", bookingID, ticketID, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// sendConfirmationEmail simulates sending a confirmation email
func sendConfirmationEmail(job BookingJob) error {
	// In production, integrate with an email service (SendGrid, SES, etc.)
	log.Printf("SIMULATED EMAIL SEND:\n"+
		"To: %s\n"+
		"Subject: Booking Confirmation #%s\n"+
		"Body: Your booking has been confirmed. %d seat(s) reserved.\n",
		job.CustomerEmail,
		job.BookingID.String()[:8],
		len(job.SeatIDs),
	)

	// Simulate network delay
	time.Sleep(100 * time.Millisecond)

	return nil
}
