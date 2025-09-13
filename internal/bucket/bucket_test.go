package bucket

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRL(t *testing.T) (*RateLimiter, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	rl := NewRateLimiter(rdb, 5*time.Minute)

	teardown := func() {
		_ = rdb.Close()
		mr.Close()
	}
	return rl, teardown
}

func TestAllowInitialAndConsume(t *testing.T) {
	ctx := context.Background()
	rl, cleanup := newTestRL(t)
	defer cleanup()

	key := "test:allow:init"
	cfg := BucketConfig{Capacity: 5, RefillPerMinute: 5} // 5 tokens, 5/min -> 1 token per 12s

	// First 5 attempts should be allowed
	for i := 0; i < 5; i++ {
		ok, rem, err := rl.Allow(ctx, key, cfg, 1)
		if err != nil {
			t.Fatalf("Allow err: %v", err)
		}
		if !ok {
			t.Fatalf("expected allowed on attempt %d, but was blocked", i+1)
		}
		if rem < 0 || rem > float64(cfg.Capacity) {
			t.Fatalf("unexpected remaining tokens: %v", rem)
		}
	}

	// 6th attempt should be blocked (no tokens left)
	ok, rem, err := rl.Allow(ctx, key, cfg, 1)
	if err != nil {
		t.Fatalf("Allow err on 6th: %v", err)
	}
	if ok {
		t.Fatalf("expected blocked on 6th attempt, but allowed")
	}
	t.Logf("6th blocked as expected, remaining tokens=%v", rem)
}

func TestRefillOverTime(t *testing.T) {
	ctx := context.Background()
	rl, cleanup := newTestRL(t)
	defer cleanup()

	key := "test:refill"
	cfg := BucketConfig{Capacity: 2, RefillPerMinute: 60} // 60/min -> 1 token/sec

	// consume all tokens
	for i := 0; i < 2; i++ {
		ok, _, err := rl.Allow(ctx, key, cfg, 1)
		if err != nil {
			t.Fatalf("Allow err: %v", err)
		}
		if !ok {
			t.Fatalf("expected allowed while draining, got blocked at %d", i+1)
		}
	}

	// immediately blocked
	ok, _, err := rl.Allow(ctx, key, cfg, 1)
	if err != nil {
		t.Fatalf("Allow err after drain: %v", err)
	}
	if ok {
		t.Fatalf("expected blocked immediately after drain")
	}

	// wait ~1.1s to let one token refill
	time.Sleep(1100 * time.Millisecond)

	ok, rem, err := rl.Allow(ctx, key, cfg, 1)
	if err != nil {
		t.Fatalf("Allow err after sleep: %v", err)
	}
	if !ok {
		t.Fatalf("expected allowed after refill, was blocked")
	}
	if rem < 0 {
		t.Fatalf("unexpected remaining tokens: %v", rem)
	}
	t.Logf("refilled and allowed, remaining tokens=%v", rem)
}

func TestCheckAllLoginPassIP(t *testing.T) {
	ctx := context.Background()
	rl, cleanup := newTestRL(t)
	defer cleanup()

	login := "alice@example.com"
	pass := "SuperSecret!"
	ip := "198.51.100.7"

	// We'll assume production CheckAll uses:
	// login: cap 3/min, pass: cap 4/min, ip: cap 10/min
	// But to be independent of internal defaults, call Allow directly for buckets as well to assert logic.
	loginKey := "bf:login:" + login
	//passKey := "bf:pass:" + hashPassword(pass)
	//ipKey := "bf:ip:" + ip

	loginCfg := BucketConfig{Capacity: 3, RefillPerMinute: 3}
	//passCfg := BucketConfig{Capacity: 4, RefillPerMinute: 4}
	//ipCfg := BucketConfig{Capacity: 10, RefillPerMinute: 10}

	// consume login capacity completely
	for i := 0; i < 3; i++ {
		ok, _, err := rl.Allow(ctx, loginKey, loginCfg, 1)
		if err != nil {
			t.Fatalf("Allow login err: %v", err)
		}
		if !ok {
			t.Fatalf("expected login allowed on attempt %d", i+1)
		}
	}
	// login now exhausted
	ok, _, err := rl.Allow(ctx, loginKey, loginCfg, 1)
	if err != nil {
		t.Fatalf("Allow login err after drain: %v", err)
	}
	if ok {
		t.Fatalf("expected login blocked after exhausting tokens")
	}

	// ensure CheckAll returns false due to login bucket
	allowed, details, err := rl.CheckAll(ctx, login, pass, ip)
	if err != nil {
		t.Fatalf("CheckAll err: %v", err)
	}
	if allowed {
		t.Fatalf("expected CheckAll to block because login bucket exhausted")
	}
	if details == nil {
		t.Fatalf("expected details map, got nil")
	}
	if details["login"] != false {
		t.Fatalf("expected login detail to be false, got true")
	}
	// ip/pass should still be allowed
	// we won't assert exact values (depends on package defaults), but ensure keys present
	if _, ok := details["pass"]; !ok {
		t.Fatalf("details missing 'pass'")
	}
	if _, ok := details["ip"]; !ok {
		t.Fatalf("details missing 'ip'")
	}
}

func newTestLimiter(t *testing.T) (*RateLimiter, func()) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run failed: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rl := NewRateLimiter(rdb, time.Minute)
	cleanup := func() {
		_ = rdb.Close()
		mr.Close()
	}
	return rl, cleanup
}

func TestTokenBucketBurstAndRefill(t *testing.T) {
	ctx := context.Background()
	rl, cleanup := newTestLimiter(t)
	defer cleanup()
	key := "test:burst"
	cfg := BucketConfig{Capacity: 10, RefillPerMinute: 10} // 10 токенов, 10 в минуту (~1 токен/6 сек)
	for i := 1; i <= 10; i++ {
		ok, _, err := rl.Allow(ctx, key, cfg, 1)
		if err != nil {
			t.Fatalf("Allow error on attempt %d: %v", i, err)
		}
		if !ok {
			t.Fatalf("expected allowed on attempt %d, got blocked", i)
		}
	}
	ok, _, err := rl.Allow(ctx, key, cfg, 1)
	if err != nil {
		t.Fatalf("Allow error on 11th attempt: %v", err)
	}
	if ok {
		t.Fatalf("expected 11th attempt to be blocked, but allowed")
	}
	time.Sleep(6 * time.Second)
	ok1, _, err := rl.Allow(ctx, key, cfg, 1)
	if err != nil {
		t.Fatalf("Allow error on first post-refill: %v", err)
	}
	if !ok1 {
		t.Fatalf("expected first post-refill request to be allowed")
	}

	ok2, _, err := rl.Allow(ctx, key, cfg, 1)
	if err != nil {
		t.Fatalf("Allow error on second post-refill: %v", err)
	}
	if ok2 {
		t.Fatalf("expected second post-refill request to be blocked, only 1 token refilled")
	}
}
