# Seat Booking System

A production-level seat booking system built with Go, featuring real-time updates via WebSockets, Redis-based seat holds, and PostgreSQL ACID transactions.

## Features

- **Session-based Seat Holding**: Redis-backed seat holds with sliding window TTL (2 minutes) and absolute session max (10 minutes)
- **ACID Booking Transactions**: PostgreSQL with `SELECT FOR UPDATE` to prevent double-bookings
- **Real-time Updates**: Redis Pub/Sub + WebSocket broadcasting for live seat status
- **Rate Limiting**: Token bucket algorithm backed by Redis (20 req/min per IP on hold endpoints)
- **Background Processing**: Async QR code generation and email notifications via Go channels
- **Graceful Shutdown**: Proper cleanup of connections and background workers
- **Swagger API Documentation**: Auto-generated API docs

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│                 │     │                 │     │                 │
│    Frontend     │────▶│   Go Server     │────▶│   PostgreSQL    │
│                 │     │                 │     │   (Bookings)    │
└─────────────────┘     └────────┬────────┘     └─────────────────┘
                                 │
        ┌────────────────────────┼────────────────────────┐
        │                        │                        │
        ▼                        ▼                        ▼
┌───────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    Redis      │     │   WebSocket     │     │   Background    │
│ (Holds/Cache) │     │      Hub        │     │    Workers      │
└───────────────┘     └─────────────────┘     └─────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional)

### Setup

1. **Start services**:

   ```bash
   docker-compose up -d postgres redis
   ```

2. **Run migrations**:

   ```bash
   goose -dir sql/migrations postgres "user=admin-user password=password host=localhost port=5432 dbname=cinema_db sslmode=disable" up
   ```

3. **Install dependencies**:

   ```bash
   go mod download
   ```

4. **Run the server**:

   ```bash
   go run ./cmd/api
   ```

5. **Access Swagger UI**: http://localhost:8080/swagger/index.html

## API Endpoints

### Hold Management

- `POST /api/v1/hold` - Hold a seat
- `DELETE /api/v1/hold` - Release a seat
- `GET /api/v1/hold/status` - Get session status
- `POST /api/v1/hold/extend` - Extend session TTL
- `GET /api/v1/hold/seats` - Get all held seats for a show

### Booking

- `POST /api/v1/bookings` - Create booking from held seats
- `GET /api/v1/bookings/{id}` - Get booking details

### Movies

- `POST /api/v1/movies` - Create movie
- `GET /api/v1/movies` - List all movies
- `GET /api/v1/movies/{id}` - Get movie details
- `PUT /api/v1/movies/{id}` - Update movie
- `DELETE /api/v1/movies/{id}` - Delete movie

### Halls

- `POST /api/v1/halls` - Create hall with seat layout
- `GET /api/v1/halls` - List all halls
- `GET /api/v1/halls/{id}` - Get hall with seats
- `DELETE /api/v1/halls/{id}` - Delete hall

### Shows

- `POST /api/v1/shows` - Create show
- `GET /api/v1/shows` - List shows (optional date filter)
- `GET /api/v1/shows/{id}` - Get show details
- `GET /api/v1/shows/{id}/seats` - Get seat availability
- `DELETE /api/v1/shows/{id}` - Delete show

### WebSocket

- `GET /ws?show_id={id}` - Connect for real-time seat updates

## Hold Flow

1. User generates a `session_id` (UUID) on the client
2. User clicks a seat → `POST /api/v1/hold` with `session_id`, `show_id`, `seat_id`
3. Server holds seat in Redis with 2-minute TTL
4. Every action resets the TTL (sliding window)
5. Session expires after 10 minutes absolute max
6. Max 15 toggles (hold/release) per session

## Booking Flow

1. User completes hold selections
2. `POST /api/v1/bookings` with `session_id`, customer details
3. Server validates held seats belong to session
4. PostgreSQL transaction with `SELECT FOR UPDATE` locks seats
5. Creates booking and tickets atomically
6. Releases Redis holds
7. Publishes WebSocket event
8. Background worker generates QR codes and sends email

## Configuration

Environment variables (see `.env`):

```env
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=admin-user
POSTGRES_PASSWORD=password
POSTGRES_DB=cinema_db

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=password

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Hold Settings (optional)
HOLD_IDLE_TTL=2m
HOLD_MAX_SESSION_TIME=10m
HOLD_MAX_TOGGLE_COUNT=15

# Rate Limiting (optional)
RATE_LIMIT_HOLD_PER_MINUTE=20
```

## Project Structure

```
├── cmd/api/              # Application entry point
├── internal/
│   ├── booking/          # Booking domain (PostgreSQL)
│   ├── cache/            # Redis client
│   ├── common/           # Shared errors, responses, validators
│   ├── config/           # Configuration
│   ├── db/               # SQLC generated code
│   ├── hold/             # Hold domain (Redis)
│   ├── middleware/       # HTTP middleware
│   ├── movies/           # Movies domain
│   ├── pubsub/           # Redis Pub/Sub
│   ├── seats/            # Seats/Halls domain
│   ├── shows/            # Shows domain
│   ├── store/            # PostgreSQL connection
│   ├── websocket/        # WebSocket hub
│   └── worker/           # Background workers
├── sql/
│   ├── migrations/       # Database migrations
│   └── queries/          # SQLC queries
└── docs/                 # Swagger docs (generated)
```

## Generate Swagger Docs

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs
```

## License

MIT
