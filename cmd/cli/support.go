package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

type response struct {
	Ok     bool   `json:"ok"`
	Reason string `json:"reason,omitempty"`
}

func doPost(url string, data any) []byte {
	var body *bytes.Reader
	if data != nil {
		jsonData, _ := json.Marshal(data)
		body = bytes.NewReader(jsonData)
	} else {
		body = bytes.NewReader([]byte{})
	}
	//nolint: gosec // reason - предположительно, что админ в курсе
	resp, err := http.Post(url, "application/json", body) //nolint: noctx
	if err != nil {
		log.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		//nolint: gocritic
		log.Fatalf("server error: %s, response: %s", resp.Status, buf.String())
	}
	return buf.Bytes()
}

func doGet(url string) []byte {
	//nolint: gosec // reason - предположительно, что админ в курсе
	resp, err := http.Get(url) //nolint: noctx
	if err != nil {
		log.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		//nolint: gocritic
		log.Fatalf("server error: %s, response: %s", resp.Status, buf.String())
	}
	return buf.Bytes()
}

func expandIPs(input string) []string {
	ips := strings.Split(input, ",")
	results := make([]string, 0, len(ips))
	for _, raw := range ips {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if strings.Contains(raw, "-") {
			parts := strings.Split(raw, "-")
			if len(parts) != 2 {
				log.Fatalf("invalid IP range: %s", raw)
			}
			start := net.ParseIP(strings.TrimSpace(parts[0]))
			end := net.ParseIP(strings.TrimSpace(parts[1]))
			if start == nil || end == nil {
				log.Fatalf("invalid IP in range: %s", raw)
			}
			for ip := start; ; ip = nextIP(ip) {
				results = append(results, ip.String())
				if ip.Equal(end) {
					break
				}
			}
			continue
		}
		if strings.Contains(raw, "/") {
			if _, _, err := net.ParseCIDR(raw); err != nil {
				log.Fatalf("invalid CIDR: %s", raw)
			}
			results = append(results, raw)
			continue
		}
		ip := net.ParseIP(raw)
		if ip == nil {
			log.Fatalf("invalid IP: %s", raw)
		}
		results = append(results, raw)
	}
	return results
}

func nextIP(ip net.IP) net.IP {
	ip = ip.To4()
	if ip == nil {
		log.Fatalf("only IPv4 supported: %s", ip)
	}
	next := make(net.IP, len(ip))
	copy(next, ip)
	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] != 0 {
			break
		}
	}
	return next
}

func resetBuckets(addr string) {
	var resp response
	if err := json.Unmarshal(doPost(addr+"/api/bucket/reset", nil), &resp); err != nil {
		fmt.Printf("❌ Failed to reset buckets: %v\n", err)
	}
	if !resp.Ok {
		if resp.Reason != "" {
			fmt.Printf("❌ Failed to reset buckets: %s\n", resp.Reason)
			return
		}
		fmt.Printf("❌ Failed to reset buckets: unknown error\n")
		return
	}
	fmt.Println("✅ All buckets successfully reset")
}

func whitelistAdd(addr string, ip string) {
	var resp response
	for _, ip := range expandIPs(ip) {
		if err := json.Unmarshal(doPost(addr+"/api/whitelist/add", map[string]string{"ip": ip}), &resp); err != nil {
			fmt.Printf("❌ Failed to whitelist %s: %v\n", ip, err)
		}
		if !resp.Ok {
			if resp.Reason != "" {
				fmt.Printf("❌ Failed to whitelist %s: %s\n", ip, resp.Reason)
				continue
			}
			fmt.Printf("❌ Failed to whitelist %s: unknown error\n", ip)
			continue
		}
		fmt.Printf("✅ Whitelisted %s\n", ip)
	}
}

func whitelistDel(addr string, ip string) {
	var resp response
	for _, ip := range expandIPs(ip) {
		if err := json.Unmarshal(doPost(addr+"/api/whitelist/del", map[string]string{"ip": ip}), &resp); err != nil {
			fmt.Printf("❌ Failed to remove %s from whitelist: %v\n", ip, err)
		}
		if !resp.Ok {
			if resp.Reason != "" {
				fmt.Printf("❌ Failed to remove %s from whitelist: %s\n", ip, resp.Reason)
				continue
			}
			fmt.Printf("❌ Failed to remove %s from whitelist: unknown error\n", ip)
			continue
		}
		doPost(addr+"/api/whitelist/del", map[string]string{"ip": ip})
		fmt.Printf("✅ Removed %s from whitelist\n", ip)
	}
}

func blacklistAdd(addr string, ip string) {
	var resp response
	for _, ip := range expandIPs(ip) {
		if err := json.Unmarshal(doPost(addr+"/api/blacklist/add", map[string]string{"ip": ip}), &resp); err != nil {
			fmt.Printf("❌ Failed to blacklist %s: %v\n", ip, err)
		}
		if !resp.Ok {
			if resp.Reason != "" {
				fmt.Printf("❌ Failed to blacklist %s: %s\n", ip, resp.Reason)
				continue
			}
			fmt.Printf("❌ Failed to blacklist %s: unknown error\n", ip)
			continue
		}
		fmt.Printf("✅ Blacklisted %s\n", ip)
	}
}

func blacklistDel(addr string, ips string) {
	var resp response
	for _, ip := range expandIPs(ips) {
		if err := json.Unmarshal(doPost(addr+"/api/blacklist/del", map[string]string{"ip": ip}), &resp); err != nil {
			fmt.Printf("❌ Failed to remove %s from blacklist: %v\n", ip, err)
		}
		if !resp.Ok {
			if resp.Reason != "" {
				fmt.Printf("❌ Failed to remove %s from blacklist: %s\n", ip, resp.Reason)
			} else {
				fmt.Printf("❌ Failed to remove %s from blacklist: unknown error\n", ip)
			}
		}
		fmt.Printf("✅ Removed %s from blacklist\n", ip)
	}
}
