package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"go-limiter/limiter"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Fatal: Failed to connect to Redis at %s: %v", redisAddr, err)
	}
	log.Printf("Successfully connected to Redis at %s", redisAddr)

	rl := limiter.NewRateLimiter(rdb)

	helloHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Request processed successfully.\n"))
	}

	http.HandleFunc("/data", rl.Middleware(5, 30*time.Second, helloHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Server is starting on port %s...", port)
	
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed: %v", err)
	}
}