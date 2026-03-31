package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shenith404/seat-booking/internal/common"
)

// RateLimiter implements a Redis-backed token bucket rate limiter
type RateLimiter struct {
	client     *redis.Client
	maxTokens  int
	refillRate time.Duration
	keyPrefix  string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *redis.Client, maxTokens int, refillRate time.Duration) *RateLimiter {
	return &RateLimiter{
		client:     client,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		keyPrefix:  "rate_limit:",
	}
}

// Middleware returns HTTP middleware that rate limits based on client IP
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		key := rl.keyPrefix + clientIP

		allowed, err := rl.allowRequest(r.Context(), key)
		if err != nil {
			// If Redis fails, allow the request but log the error
			fmt.Printf("Rate limiter error: %v\n", err)
			next.ServeHTTP(w, r)
			return
		}

		if !allowed {
			w.Header().Set("Retry-After", fmt.Sprintf("%d", int(rl.refillRate.Seconds())))
			common.Err(w, common.NewRateLimitError())
			return
		}

		next.ServeHTTP(w, r)
	})
}

// allowRequest checks if a request should be allowed using token bucket algorithm
func (rl *RateLimiter) allowRequest(ctx context.Context, key string) (bool, error) {
	// Lua script for atomic token bucket operation
	script := redis.NewScript(`
		local key = KEYS[1]
		local max_tokens = tonumber(ARGV[1])
		local refill_time = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local tokens = tonumber(bucket[1])
		local last_refill = tonumber(bucket[2])

		-- Initialize bucket if it doesn't exist
		if tokens == nil then
			tokens = max_tokens
			last_refill = now
		end

		-- Calculate tokens to add based on elapsed time
		local elapsed = now - last_refill
		local tokens_to_add = math.floor(elapsed / refill_time)
		tokens = math.min(max_tokens, tokens + tokens_to_add)

		-- Update last refill time if tokens were added
		if tokens_to_add > 0 then
			last_refill = now
		end

		-- Try to consume a token
		if tokens > 0 then
			tokens = tokens - 1
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', last_refill)
			redis.call('EXPIRE', key, refill_time * max_tokens * 2)
			return 1
		else
			redis.call('HMSET', key, 'tokens', tokens, 'last_refill', last_refill)
			redis.call('EXPIRE', key, refill_time * max_tokens * 2)
			return 0
		end
	`)

	now := time.Now().Unix()
	refillSeconds := int64(rl.refillRate.Seconds())
	if refillSeconds == 0 {
		refillSeconds = 1
	}

	result, err := script.Run(ctx, rl.client, []string{key}, rl.maxTokens, refillSeconds, now).Int()
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
