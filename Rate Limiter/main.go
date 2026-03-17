package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"ratelimiter/config"
	"ratelimiter/limiter"
	"syscall"
	"time"
)

func serveRequest(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello world!"))
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ratelimiter, err := limiter.NewRateLimiterService(ctx, limiter.TokenBucketAlgo, config.RateLimiterConfig{
		Requests: 3,
		Window: 60 * time.Second,
		Burst: 3,
		CleanupInterval: 5 * time.Minute,
		BucketTTL: 10 * time.Minute,
	})
	if err != nil {
		log.Fatalf("failed to create rate limiter: %v", err)
	}

	mux := http.NewServeMux()
	handler := http.HandlerFunc(serveRequest)

	mux.Handle("/", ratelimiter.HandleRequest(handler)) // wrap the handler in ratelimit middleware

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}