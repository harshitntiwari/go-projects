package limiter

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"ratelimiter/config"
	"strings"
)

type AlgoType int

const (
	DefaultAlgo AlgoType = iota
	TokenBucketAlgo
)

type RateLimiterService struct {
	rl RateLimiter
}

// allocates and returns a new [RateLimiterService]
func NewRateLimiterService(ctx context.Context, algo AlgoType, cfg config.RateLimiterConfig) (*RateLimiterService, error) {
	if algo == DefaultAlgo {
		algo = TokenBucketAlgo
	}

	var ratelimiter RateLimiter

	switch algo {
	case TokenBucketAlgo:
		ratelimiter = NewTokenBucketRateLimiter(ctx, cfg)
	default:
		return nil, fmt.Errorf("Unknown algo type: %d", algo)
	}

	return &RateLimiterService{
		rl: ratelimiter,
	}, nil
}

func (svc *RateLimiterService) HandleRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIp := getClientIP(r)
		if !svc.allowRequest(clientIp) {
			http.Error(w, "Too many requests", http.StatusTooManyRequests)
    		return
		}
		next.ServeHTTP(w, r)
	})
}

// returns true if the request satisfies the set rate limit, false otherwise
func (svc *RateLimiterService) allowRequest(clientID string) bool {
	return svc.rl.Allow(clientID)
}

func getClientIP(r *http.Request) string {
	// Check standard headers for the client's real IP, especially if behind a proxy or load balancer.
	// X-Forwarded-For can contain a list of IPs; the first one is the original client IP.
	// X-Real-IP is also commonly used.
	for _, headerName := range []string{"X-Forwarded-For", "X-Real-IP"} {
		headerValue := r.Header.Get(headerName)
		if headerValue != "" {
			ips := strings.Split(headerValue, ",")
			// The first IP in the list is typically the original client IP
			for _, ipStr := range ips {
				ipStr = strings.TrimSpace(ipStr)
				if ipStr != "" {
					return ipStr
				}
			}
		}
	}

	// Fallback to RemoteAddr if headers are not present.
	// RemoteAddr contains both IP and port (e.g., "127.0.0.1:8080").
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return ip
	}

	return r.RemoteAddr
}