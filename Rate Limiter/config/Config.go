package config

import "time"

type RateLimiterConfig struct {
	// number of requests to allow within the window
	Requests  		int
	// window size
	Window    		time.Duration
	// max number of requests that can go in a burst
	Burst     		int
	// For TokenBucketAlgo: stale buckets will be cleaned up after every CleanupInterval
	CleanupInterval time.Duration
	// For TokenBucketAlgo: bucket will be expired BucketTTL duration after its lastAccessTime
	BucketTTL       time.Duration
}