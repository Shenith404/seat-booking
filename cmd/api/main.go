package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/shenith404/seat-booking/docs" // Swagger docs

	"github.com/google/uuid"
	"github.com/shenith404/seat-booking/internal/booking"
	"github.com/shenith404/seat-booking/internal/cache"
	"github.com/shenith404/seat-booking/internal/config"
	"github.com/shenith404/seat-booking/internal/hold"
	"github.com/shenith404/seat-booking/internal/middleware"
	"github.com/shenith404/seat-booking/internal/movies"
	"github.com/shenith404/seat-booking/internal/pubsub"
	"github.com/shenith404/seat-booking/internal/seats"
	"github.com/shenith404/seat-booking/internal/shows"
	"github.com/shenith404/seat-booking/internal/store"
	"github.com/shenith404/seat-booking/internal/websocket"
	"github.com/shenith404/seat-booking/internal/worker"
)

// @title Seat Booking API
// @version 1.0
// @description Production-level seat booking system with real-time updates
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-Session-ID

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize PostgreSQL
	pgStore, err := store.NewPostgresStore(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgStore.Close()

	// Initialize Redis
	redisCache, err := cache.NewRedisCache(ctx, cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	// Initialize PubSub
	ps := pubsub.NewPubSub(redisCache.Client)

	// Initialize WebSocket Hub
	wsHub := websocket.NewHub(ps)
	go wsHub.Run(ctx)

	// Initialize Background Worker
	bgWorker := worker.NewWorker(3, 100)
	bgWorker.Start(ctx)
	defer bgWorker.Stop()

	// Initialize Repositories
	movieRepo := movies.NewPostgresRepository(pgStore.Pool)
	seatRepo := seats.NewPostgresRepository(pgStore.Pool)
	showRepo := shows.NewPostgresRepository(pgStore.Pool)
	bookingRepo := booking.NewPostgresRepository(pgStore.Pool)
	holdRepo := hold.NewRedisRepository(redisCache.Client)

	// Initialize Services
	holdService := hold.NewService(holdRepo, ps, hold.ServiceConfig{
		IdleTTL:        cfg.Hold.IdleTTL,
		MaxSessionTime: cfg.Hold.MaxSessionTime,
		MaxToggleCount: cfg.Hold.MaxToggleCount,
	})
	movieService := movies.NewService(movieRepo)
	seatService := seats.NewService(seatRepo)
	showService := shows.NewService(showRepo, movieRepo, holdService)
	bookingService := booking.NewService(bookingRepo, holdService, ps, bgWorker)

	// Set QR callback for worker
	bgWorker.SetQRCallback(func(ctx context.Context, ticketID uuid.UUID, qrHash string) error {
		return bookingRepo.UpdateTicketQRHash(ctx, ticketID, qrHash)
	})

	// Initialize Handlers
	holdHandler := hold.NewHandler(holdService)
	movieHandler := movies.NewHandler(movieService)
	seatHandler := seats.NewHandler(seatService)
	showHandler := shows.NewHandler(showService)
	bookingHandler := booking.NewHandler(bookingService)
	wsHandler := websocket.NewHandler(wsHub)

	// Initialize Rate Limiter
	rateLimiter := middleware.NewRateLimiter(
		redisCache.Client,
		cfg.RateLimit.HoldRequestsPerMinute,
		cfg.RateLimit.BucketTTL,
	)

	// Setup Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recovery)
	r.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// WebSocket endpoint
	r.Get("/ws", wsHandler.ServeWS)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Hold routes with rate limiting
		r.Group(func(r chi.Router) {
			r.Use(rateLimiter.Middleware)
			holdHandler.RegisterRoutes(r)
		})

		// Booking routes
		bookingHandler.RegisterRoutes(r)

		// Movie routes
		movieHandler.RegisterRoutes(r)

		// Hall/Seat routes
		seatHandler.RegisterRoutes(r)

		// Show routes
		showHandler.RegisterRoutes(r)
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting server on %s:%d", cfg.Server.Host, cfg.Server.Port)
		log.Printf("Swagger UI available at http://localhost:%d/swagger/index.html", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Cancel context to stop background goroutines
	cancel()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped gracefully")
}
