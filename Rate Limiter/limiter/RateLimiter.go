package limiter

import (
	"context"
	"fmt"
	"ratelimiter/config"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(clientID string) bool
}

type TokenBucket struct {
	tokens         float64
	lastRefillTime time.Time
	lastAccessTime time.Time
	mu             sync.Mutex
}

// allocates and returns a new [TokenBucket]
func NewTokenBucket(initialTokens float64) *TokenBucket {
	return &TokenBucket{
		tokens: initialTokens,
		lastRefillTime: time.Now(),
	}
}

// refills the bucket and returns true by consuming 1 token if available, else false
func (b *TokenBucket) refillAndConsume(refillRate float64, capacity float64) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	elapsedTime := time.Since(b.lastRefillTime) // same as time.Now().Sub(b.lastRefillTime)

	tokensToBeAdded  := elapsedTime.Seconds() * refillRate

	b.tokens  = min(b.tokens + tokensToBeAdded, capacity)

	// update last refill time to current time
	b.lastRefillTime = time.Now() 

	// if 1 complete token is not available, return false
	if b.tokens < 0 {
		return false
	}

	// consume 1 token
	b.tokens -= 1
	return true
}

type TokenBucketRateLimiter struct{
	buckets map[string]*TokenBucket
	config  config.RateLimiterConfig
	mu      sync.RWMutex
}

// allocates and returns a new [TokenBucketRateLimiter]
func NewTokenBucketRateLimiter(ctx context.Context, cfg config.RateLimiterConfig) *TokenBucketRateLimiter {
	rl := &TokenBucketRateLimiter{
		buckets: map[string]*TokenBucket{},
		config: cfg,
	}

	// spawn a cleanup worker
	go rl.cleanupWorker(ctx)

	return rl
}

// worker wakes up at a fixed interval, cleans up the expired buckets, then goes to sleep
func (rl *TokenBucketRateLimiter) cleanupWorker(ctx context.Context) {
	ticker := time.NewTicker(rl.config.CleanupInterval)

	for {
		select {
		case <- ctx.Done():
			ticker.Stop()
			fmt.Printf("graceful shutdown: %v", ctx.Err())
			return
		case <- ticker.C: // triggers after every CleanupInterval
			rl.mu.Lock()
			for clientID, bucket := range rl.buckets {
				// if time elapsed since last access is more than the bucket TTL, delete the bucket
				bucket.mu.Lock()
				isStale := time.Since(bucket.lastAccessTime) > rl.config.BucketTTL
				bucket.mu.Unlock()
				if  isStale {
					// delete the bucket
					delete(rl.buckets, clientID)
				}
			}
			rl.mu.Unlock()
		}
	}	
}

func (rl * TokenBucketRateLimiter) getBucket(clientID string) *TokenBucket {
	rl.mu.RLock()
	bucket, ok := rl.buckets[clientID]; 
	rl.mu.RUnlock()
	if ok {
		return bucket
	}

	// create the bucket if not already created for this client
	rl.mu.Lock()
	defer rl.mu.Unlock()
	// checking again b'cz another goroutine could've created the bucket b/w this goroutine's RUnlock() and Lock()
	if bucket, ok := rl.buckets[clientID]; ok {
		return bucket
	}
	// initially we're keeping the bucket full
	rl.buckets[clientID] = NewTokenBucket(float64(rl.config.Burst))

	return rl.buckets[clientID]
}

func (rl *TokenBucketRateLimiter) Allow(clientID string) bool {
	bucket := rl.getBucket(clientID)

	refillRate := float64(rl.config.Requests) / rl.config.Window.Seconds()

	return bucket.refillAndConsume(float64(refillRate), float64(rl.config.Burst))
}