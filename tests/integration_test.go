//go:build ignore

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"testing"
)

const baseURL = "http://localhost:8888"

type AuthorizeResp struct {
	Ok bool `json:"ok"`
}

func postAuthorize(t *testing.T, body any) bool {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := http.Post(baseURL+"/api/authorize", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/authorize failed: %v", err)
	}
	// nolint: errcheck
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	var res AuthorizeResp
	if err := json.Unmarshal(data, &res); err != nil {
		t.Fatalf("invalid JSON response: %s", string(data))
	}
	return res.Ok
}

func postWhitelist(t *testing.T, ip string) {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"ip": ip})
	resp, err := http.Post(baseURL+"/api/whitelist/add", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST /api/whitelist failed: %v", err)
	}
	// nolint: errcheck
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("whitelist request failed, code=%d", resp.StatusCode)
	}
}

func postResetBucket(t *testing.T) {
	t.Helper()
	resp, err := http.Post(baseURL+"/api/bucket/reset", "application/json", nil)
	if err != nil {
		t.Fatal("POST /api/bucket/reset failed:", err)
	}
	//nolint: errcheck
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		t.Fatalf(
			"reset bucket request failed, code=%d, text=%v",
			resp.StatusCode,
			string(body),
		)
	}
}

func postResetBucketLogin(t *testing.T, login string) {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"login": login})
	resp, err := http.Post(baseURL+"/api/bucket/reset/login", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal("POST /api/bucket/reset failed:", err)
	}
	//nolint: errcheck
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf(
			"reset bucket request failed, code=%d",
			resp.StatusCode,
		)
	}
}

func postResetBucketIP(t *testing.T, IP string) {
	t.Helper()
	b, _ := json.Marshal(map[string]string{"ip": IP})
	resp, err := http.Post(baseURL+"/api/bucket/reset/ip", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal("POST /api/bucket/reset failed:", err)
	}
	// nolint: errcheck
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf(
			"reset bucket request failed, code=%d",
			resp.StatusCode,
		)
	}
}

func TestRateLimiterBlocks(t *testing.T) {
	login := "user@example.com"
	pass := "pass1"
	postResetBucket(t)
	for i := 1; i <= 10; i++ {
		ok := postAuthorize(t, map[string]string{
			"login":    login,
			"password": fmt.Sprintf("%s%d", pass, i),
			"ip":       fmt.Sprintf("192.168.1.%d", i),
		})
		if !ok {
			t.Fatalf("expected allowed on attempt %d", i)
		}
	}
	ok := postAuthorize(t, map[string]string{
		"login":    login,
		"password": "anotherpass",
		"ip":       "192.168.1.100",
	})
	if ok {
		t.Fatalf("expected blocked on 11th attempt")
	}
	postResetBucketLogin(t, login)
	for i := 1; i <= 10; i++ {
		ok := postAuthorize(t, map[string]string{
			"login":    login,
			"password": fmt.Sprintf("%s%d", pass, i),
			"ip":       fmt.Sprintf("192.168.1.%d", i),
		})
		if !ok {
			t.Fatalf("expected allowed on attempt %d", i)
		}
	}
	ok = postAuthorize(t, map[string]string{
		"login":    login,
		"password": "anotherpass",
		"ip":       "192.168.1.100",
	})
	if ok {
		t.Fatalf("expected blocked on 11th attempt")
	}
	pass = "secretpass"
	for i := 1; i <= 100; i++ {
		ok := postAuthorize(t, map[string]string{
			"login":    fmt.Sprintf("user%d@example.com", i),
			"password": pass,
			"ip":       fmt.Sprintf("10.0.0.%d", i),
		})
		if !ok {
			t.Fatalf("expected allowed on password attempt %d", i)
		}
	}
	ok = postAuthorize(t, map[string]string{
		"login":    "anotheruser@example.com",
		"password": pass,
		"ip":       "10.0.0.200",
	})
	if ok {
		t.Fatalf("expected password blocked on 101st attempt")
	}
	ip := "203.0.113.1"
	first_run := 0
	for i := 1; i <= 2000; i++ {
		ok := postAuthorize(t, map[string]string{
			"login":    fmt.Sprintf("user%d@example.com", i),
			"password": fmt.Sprintf("pass%d", i),
			"ip":       ip,
		})
		if !ok {
			first_run = i
			break
		}
	}
	postResetBucketIP(t, ip)
	for i := 1; i <= 2000; i++ {
		ok := postAuthorize(t, map[string]string{
			"login":    fmt.Sprintf("user%d@example.com", i),
			"password": fmt.Sprintf("pass%d", i),
			"ip":       ip,
		})
		if !ok {
			if math.Abs(float64(i-first_run)) > 5 {
				t.Fatalf(
					"expected allowed on ip reset, got blocked on attempt First run: %d Second run: %d",
					first_run, i,
				)
			}
			return
		}
	}
	t.Fatal("Ip limit exceeded")
}

func TestWhitelistBypassesRateLimit(t *testing.T) {
	postResetBucketLogin(t, "whitelistuser")
	ip := "198.51.100.7"
	login := "whitelistuser"
	postWhitelist(t, ip)
	pass := "somepass"
	for i := 1; i <= 50; i++ {
		ok := postAuthorize(t, map[string]string{
			"login":    login,
			"password": pass,
			"ip":       ip,
		})
		if !ok {
			t.Fatalf("expected allowed due to whitelist, got blocked on attempt %d", i)
		}
	}
}
