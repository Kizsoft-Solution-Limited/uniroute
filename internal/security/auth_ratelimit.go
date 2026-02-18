package security

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/Kizsoft-Solution-Limited/uniroute/internal/storage"
	"github.com/redis/go-redis/v9"
)

type AuthRateLimiter struct {
	redis *storage.RedisClient
}

func NewAuthRateLimiter(redis *storage.RedisClient) *AuthRateLimiter {
	return &AuthRateLimiter{
		redis: redis,
	}
}

// identifier can be IP-based ("ip:...") or email-based ("email:..."). Returns: allowed, waitTime (seconds), blocked, error.
// - waitTime: how long to wait before next attempt (if not allowed)
// - blocked: whether the account is temporarily blocked
func (r *AuthRateLimiter) CheckAuthRateLimit(ctx context.Context, identifier string, maxAttempts int) (allowed bool, waitTime int, blocked bool, err error) {
	client := r.redis.Client()

	attemptsKey := fmt.Sprintf("auth:attempts:%s", identifier)
	blockedKey := fmt.Sprintf("auth:blocked:%s", identifier)

	blockedUntil, err := client.Get(ctx, blockedKey).Int64()
	if err == nil && blockedUntil > time.Now().Unix() {
		waitTime = int(blockedUntil - time.Now().Unix())
		return false, waitTime, true, nil
	}

	attempts, err := client.Get(ctx, attemptsKey).Int()
	if err == redis.Nil {
		attempts = 0
	} else if err != nil {
		return false, 0, false, fmt.Errorf("failed to get attempt count: %w", err)
	}

	if attempts >= maxAttempts {
		blockUntil := time.Now().Add(15 * time.Minute).Unix()
		client.Set(ctx, blockedKey, blockUntil, 15*time.Minute)
		client.Del(ctx, attemptsKey) // Reset attempts after blocking
		waitTime = 15 * 60           // 15 minutes in seconds
		return false, waitTime, true, nil
	}

	if attempts > 0 {
		waitTime = int(math.Pow(2, float64(attempts)))
		if waitTime > 60 {
			waitTime = 60
		}
	}

	lastAttemptKey := fmt.Sprintf("auth:last_attempt:%s", identifier)
	lastAttempt, err := client.Get(ctx, lastAttemptKey).Int64()
	if err == nil && lastAttempt > 0 {
		timeSinceLastAttempt := time.Now().Unix() - lastAttempt
		if int(timeSinceLastAttempt) < waitTime {
			remainingWait := waitTime - int(timeSinceLastAttempt)
			return false, remainingWait, false, nil
		}
	}

	// Request is allowed
	return true, 0, false, nil
}

func (r *AuthRateLimiter) RecordFailedAttempt(ctx context.Context, identifier string) error {
	client := r.redis.Client()

	attemptsKey := fmt.Sprintf("auth:attempts:%s", identifier)
	lastAttemptKey := fmt.Sprintf("auth:last_attempt:%s", identifier)

	// Increment attempt count
	attempts, err := client.Incr(ctx, attemptsKey).Result()
	if err != nil {
		return fmt.Errorf("failed to increment attempts: %w", err)
	}

	if attempts == 1 {
		client.Expire(ctx, attemptsKey, 24*time.Hour)
	}

	client.Set(ctx, lastAttemptKey, time.Now().Unix(), 24*time.Hour)

	return nil
}

func (r *AuthRateLimiter) RecordSuccess(ctx context.Context, identifier string) error {
	client := r.redis.Client()

	attemptsKey := fmt.Sprintf("auth:attempts:%s", identifier)
	lastAttemptKey := fmt.Sprintf("auth:last_attempt:%s", identifier)
	blockedKey := fmt.Sprintf("auth:blocked:%s", identifier)

	client.Del(ctx, attemptsKey)
	client.Del(ctx, lastAttemptKey)
	client.Del(ctx, blockedKey)

	return nil
}

func (r *AuthRateLimiter) GetWaitTime(ctx context.Context, identifier string) (waitTime int, blocked bool, err error) {
	client := r.redis.Client()

	blockedKey := fmt.Sprintf("auth:blocked:%s", identifier)
	attemptsKey := fmt.Sprintf("auth:attempts:%s", identifier)
	lastAttemptKey := fmt.Sprintf("auth:last_attempt:%s", identifier)

	blockedUntil, err := client.Get(ctx, blockedKey).Int64()
	if err == nil && blockedUntil > time.Now().Unix() {
		waitTime = int(blockedUntil - time.Now().Unix())
		return waitTime, true, nil
	}

	attempts, err := client.Get(ctx, attemptsKey).Int()
	if err == redis.Nil {
		return 0, false, nil
	} else if err != nil {
		return 0, false, err
	}

	if attempts == 0 {
		return 0, false, nil
	}

	waitTime = int(math.Pow(2, float64(attempts)))
	if waitTime > 60 {
		waitTime = 60
	}

	lastAttempt, err := client.Get(ctx, lastAttemptKey).Int64()
	if err == nil && lastAttempt > 0 {
		timeSinceLastAttempt := time.Now().Unix() - lastAttempt
		if int(timeSinceLastAttempt) < waitTime {
			remainingWait := waitTime - int(timeSinceLastAttempt)
			return remainingWait, false, nil
		}
	}

	return 0, false, nil
}
