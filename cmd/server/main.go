package main

import (
	"log"
	"time"

	"github.com/meladark/special-train/configs"
	"github.com/meladark/special-train/internal/api"
	"github.com/meladark/special-train/internal/app"
	"github.com/meladark/special-train/internal/bucket"
	"github.com/meladark/special-train/internal/service"
	"github.com/meladark/special-train/internal/storage"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := configs.LoadConfig()
	store := storage.NewInMemoryStorage()
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
		DB:   0,
	})
	defer rdb.Close()
	rl := bucket.NewRateLimiter(rdb, 5*time.Minute,
		bucket.BucketConfig{Capacity: cfg.CLogin, RefillPerMinute: cfg.RLogin},
		bucket.BucketConfig{Capacity: cfg.CPass, RefillPerMinute: cfg.RPass},
		bucket.BucketConfig{Capacity: cfg.CIP, RefillPerMinute: cfg.RIP})
	svc := service.New(store, rl)
	router := api.NewRouter(svc)
	srv := app.NewServer(":"+cfg.Port, router)
	if err := srv.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
