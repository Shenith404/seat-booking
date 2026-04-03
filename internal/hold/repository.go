package hold

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Repository defines the interface for hold data access
type Repository interface {
	// Session operations
	GetSession(ctx context.Context, showID, sessionID string) (*Session, error)
	SaveSession(ctx context.Context, session *Session, ttl time.Duration) error
	DeleteSession(ctx context.Context, showID, sessionID string) error
	ExtendSession(ctx context.Context, showID, sessionID string, ttl time.Duration) error

	// Seat hold operations
	HoldSeat(ctx context.Context, showID, seatID, sessionID string, ttl time.Duration) (bool, error)
	ExtendSeatHold(ctx context.Context, showID, seatID, sessionID string, ttl time.Duration) error
	ReleaseSeat(ctx context.Context, showID, seatID, sessionID string) error
	GetSeatHolder(ctx context.Context, showID, seatID string) (string, error)
	GetHeldSeats(ctx context.Context, showID string) (map[string]string, error)
	ReleaseAllSeats(ctx context.Context, showID, sessionID string, seatIDs []string) error
}

// RedisRepository implements Repository using Redis
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new Redis repository
func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

// Key generators
func sessionKey(showID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", showID, sessionID)
}

func seatHoldKey(showID, seatID string) string {
	return fmt.Sprintf("seat_hold:%s:%s", showID, seatID)
}

func showSeatsKey(showID string) string {
	return fmt.Sprintf("show_seats:%s", showID)
}

// GetSession retrieves a session from Redis
func (r *RedisRepository) GetSession(ctx context.Context, showID, sessionID string) (*Session, error) {
	key := sessionKey(showID, sessionID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session: %w", err)
	}

	return &session, nil
}

// SaveSession saves a session to Redis
func (r *RedisRepository) SaveSession(ctx context.Context, session *Session, ttl time.Duration) error {
	key := sessionKey(session.ShowID, session.SessionID)
	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	return nil
}

// DeleteSession deletes a session from Redis
func (r *RedisRepository) DeleteSession(ctx context.Context, showID, sessionID string) error {
	key := sessionKey(showID, sessionID)
	return r.client.Del(ctx, key).Err()
}

// ExtendSession extends the TTL of a session
func (r *RedisRepository) ExtendSession(ctx context.Context, showID, sessionID string, ttl time.Duration) error {
	key := sessionKey(showID, sessionID)
	return r.client.Expire(ctx, key, ttl).Err()
}

// HoldSeat attempts to hold a seat for a session using SET with NX option
func (r *RedisRepository) HoldSeat(ctx context.Context, showID, seatID, sessionID string, ttl time.Duration) (bool, error) {
	key := seatHoldKey(showID, seatID)

	// Use SET with NX option to atomically set if not exists (redis v9 uses SetArgs)
	success, err := r.client.SetArgs(ctx, key, sessionID, redis.SetArgs{
		Mode: "NX",
		TTL:  ttl,
	}).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to hold seat: %w", err)
	}

	if success == "OK" {
		// Add to show's held seats set
		r.client.SAdd(ctx, showSeatsKey(showID), seatID)
		r.client.Expire(ctx, showSeatsKey(showID), ttl*10) // 10x seat hold time

		return true, nil
	}

	return false, nil
}

// ExtendSeatHold extends the TTL of an existing seat hold (only if owned by the session)
func (r *RedisRepository) ExtendSeatHold(ctx context.Context, showID, seatID, sessionID string, ttl time.Duration) error {
	key := seatHoldKey(showID, seatID)

	// Use Lua script to atomically check ownership and extend TTL
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		end
		return 0
	`)

	_, err := script.Run(ctx, r.client, []string{key}, sessionID, int(ttl.Seconds())).Result()
	if err != nil {
		return fmt.Errorf("failed to extend seat hold: %w", err)
	}

	return nil
}

// ReleaseSeat releases a seat hold
func (r *RedisRepository) ReleaseSeat(ctx context.Context, showID, seatID, sessionID string) error {
	key := seatHoldKey(showID, seatID)

	// Only delete if the session owns the seat
	script := redis.NewScript(`
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		end
		return 0
	`)

	_, err := script.Run(ctx, r.client, []string{key}, sessionID).Result()
	if err != nil {
		return fmt.Errorf("failed to release seat: %w", err)
	}

	// Remove from show's held seats set
	r.client.SRem(ctx, showSeatsKey(showID), seatID)

	return nil
}

// GetSeatHolder returns the session ID holding a seat, or empty string
func (r *RedisRepository) GetSeatHolder(ctx context.Context, showID, seatID string) (string, error) {
	key := seatHoldKey(showID, seatID)
	holder, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("failed to get seat holder: %w", err)
	}
	return holder, nil
}

// GetHeldSeats returns all held seats for a show (seatID -> sessionID)
func (r *RedisRepository) GetHeldSeats(ctx context.Context, showID string) (map[string]string, error) {
	// Get all seat IDs from the show's set
	seatIDs, err := r.client.SMembers(ctx, showSeatsKey(showID)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get held seats: %w", err)
	}

	result := make(map[string]string)
	for _, seatID := range seatIDs {
		holder, err := r.GetSeatHolder(ctx, showID, seatID)
		if err != nil {
			continue
		}
		if holder != "" {
			result[seatID] = holder
		} else {
			// Clean up stale entry
			r.client.SRem(ctx, showSeatsKey(showID), seatID)
		}
	}

	return result, nil
}

// ReleaseAllSeats releases multiple seats for a session
func (r *RedisRepository) ReleaseAllSeats(ctx context.Context, showID, sessionID string, seatIDs []string) error {
	for _, seatID := range seatIDs {
		if err := r.ReleaseSeat(ctx, showID, seatID, sessionID); err != nil {
			return err
		}
	}
	return nil
}
