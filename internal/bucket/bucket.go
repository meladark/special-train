package bucket

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Capacity        int
	RefillPerMinute int
}

type RateLimiter struct {
	rdb        *redis.Client
	keyTTL     time.Duration
	maxRetries int
	loginCfg   Config
	passCfg    Config
	ipCfg      Config
}

func NewRateLimiter(rdb *redis.Client,
	keyTTL time.Duration,
	loginCfg Config,
	passCfg Config,
	ipCfg Config,
) *RateLimiter {
	return &RateLimiter{
		rdb:        rdb,
		keyTTL:     keyTTL,
		maxRetries: 5,
		loginCfg:   loginCfg,
		passCfg:    passCfg,
		ipCfg:      ipCfg,
	}
}

func hashPassword(pw string) string {
	h := sha256.Sum256([]byte(pw))
	return hex.EncodeToString(h[:])
}

func (rl *RateLimiter) Allow(ctx context.Context, //nolint: gocognit
	key string,
	cfg Config,
	requested int,
) (bool, float64, error) {
	refillPerSec := float64(cfg.RefillPerMinute) / 60.0
	attempts := 0
	for attempts < rl.maxRetries {
		attempts++
		var allowed bool
		var remaining float64
		err := rl.rdb.Watch(ctx, func(tx *redis.Tx) error {
			tokensStr, err := tx.HGet(ctx, key, "tokens").Result()
			if err != nil && !errors.Is(err, redis.Nil) {
				return err
			}
			tsStr, err2 := tx.HGet(ctx, key, "ts").Result()
			if err2 != nil && !errors.Is(err2, redis.Nil) {
				return err2
			}
			now := float64(time.Now().UnixNano()) / 1e9
			var tokens float64
			if errors.Is(err, redis.Nil) {
				tokens = float64(cfg.Capacity)
			} else {
				tokens, err = strconv.ParseFloat(tokensStr, 64)
				if err != nil {
					tokens = float64(cfg.Capacity)
				}
			}
			var lastTs float64
			if errors.Is(err2, redis.Nil) {
				lastTs = now
			} else {
				lastTs, err2 = strconv.ParseFloat(tsStr, 64)
				if err2 != nil {
					lastTs = now
				}
			}
			delta := now - lastTs
			if delta < 0 {
				delta = 0
			}
			newTokens := math.Min(float64(cfg.Capacity), tokens+delta*refillPerSec)
			if newTokens >= float64(requested) {
				newTokens -= float64(requested)
				allowed = true
			} else {
				allowed = false
			}
			remaining = newTokens
			_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
				pipe.HSet(ctx, key, "tokens", strconv.FormatFloat(newTokens, 'f', 6, 64))
				pipe.HSet(ctx, key, "ts", strconv.FormatFloat(now, 'f', 6, 64))
				pipe.Expire(ctx, key, rl.keyTTL)
				return nil
			})
			return err
		}, key)
		if err == nil {
			return allowed, remaining, nil
		}
		if errors.Is(err, redis.TxFailedErr) {
			continue
		}
		return false, 0, err
	}
	return false, 0, fmt.Errorf("rate limiter: max retries reached")
}

func (rl *RateLimiter) CheckAll(ctx context.Context, login, password, ip string) (bool, map[string]bool, error) {
	passHash := hashPassword(password)

	loginKey := "bf:login:" + login
	passKey := "bf:pass:" + passHash
	ipKey := "bf:ip:" + ip

	res := map[string]bool{}

	okLogin, _, err := rl.Allow(ctx, loginKey, rl.loginCfg, 1)
	if err != nil {
		return false, nil, err
	}
	res["login"] = okLogin

	okPass, _, err := rl.Allow(ctx, passKey, rl.passCfg, 1)
	if err != nil {
		return false, nil, err
	}
	res["pass"] = okPass

	okIP, _, err := rl.Allow(ctx, ipKey, rl.ipCfg, 1)
	if err != nil {
		return false, nil, err
	}
	res["ip"] = okIP

	return okLogin && okPass && okIP, res, nil
}

func (rl *RateLimiter) ResetIP(ctx context.Context, ip string) error {
	return rl.ResetAll(ctx, "ip:"+ip)
}

func (rl *RateLimiter) ResetLogin(ctx context.Context, login string) error {
	return rl.ResetAll(ctx, "login:"+login)
}

func (rl *RateLimiter) ResetAll(ctx context.Context, reset string) error {
	iter := rl.rdb.Scan(ctx, 0, "bf:"+reset, 100).Iterator()
	for iter.Next(ctx) {
		if err := rl.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}
