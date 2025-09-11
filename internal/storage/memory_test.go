package storage

import (
	"fmt"
	"net"
	"testing"
)

func mustCIDR(s string) net.IPNet {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return *ipnet
}

func TestAddToWhitelist(t *testing.T) {
	s := NewInMemoryStorage()
	ip := mustCIDR("192.168.1.0/24")
	ok, err := s.AddToWhitelist(ip, false)
	if !ok || err != nil {
		t.Fatalf("expected success, got ok=%v err=%v", ok, err)
	}
	_, err = s.AddToWhitelist(ip, false)
	if err == nil || err.Error() != fmt.Sprintf("IP fully overlaps in whitelist: %s", ip.String()) {
		t.Errorf("expected full overlap error, got %v", err)
	}
	ip2 := mustCIDR("192.168.1.128/25")
	ok, err = s.AddToWhitelist(ip2, false)
	if ok || err == nil || err.Error() != fmt.Sprintf("IP fully overlaps in whitelist: %s", ip.String()) {
		t.Errorf("expected full overlap (subset), got ok=%v err=%v", ok, err)
	}
	if _, found := s.whitelist[ip2.String()]; found {
		t.Errorf("ip2 should NOT be added to whitelist")
	}
	ip3 := mustCIDR("10.0.0.0/8")
	s.blacklist[ip3.String()] = &ip3
	_, err = s.AddToWhitelist(ip3, false)
	if err == nil || err.Error() != fmt.Sprintf("IP overlap in blacklist: %s", ip3.String()) {
		t.Errorf("expected blacklist overlap error, got %v", err)
	}
	ok, err = s.AddToWhitelist(ip3, true)
	if !ok || err != nil {
		t.Errorf("expected force insert success, got ok=%v err=%v", ok, err)
	}
}

func TestAddToBlacklist(t *testing.T) {
	s := NewInMemoryStorage()

	// 1. Добавляем сеть в blacklist
	ip := mustCIDR("172.16.0.0/16")
	ok, err := s.AddToBlacklist(ip, false)
	if !ok || err != nil {
		t.Fatalf("expected success, got ok=%v err=%v", ok, err)
	}

	// 2. Повторное добавление → полный overlap
	_, err = s.AddToBlacklist(ip, false)
	if err == nil || err.Error() != fmt.Sprintf("IP fully overlaps in blacklist: %s", ip.String()) {
		t.Errorf("expected full overlap error, got %v", err)
	}

	// 3. Добавляем подсеть, которая вложена в существующую (/17 внутри /16) → полный overlap
	ip2 := mustCIDR("172.16.128.0/17")
	ok, err = s.AddToBlacklist(ip2, false)
	if ok || err == nil || err.Error() != fmt.Sprintf("IP fully overlaps in blacklist: %s", ip.String()) {
		t.Errorf("expected full overlap (subset), got ok=%v err=%v", ok, err)
	}
	if _, found := s.blacklist[ip2.String()]; found {
		t.Errorf("ip2 should NOT be added to blacklist")
	}

	// 4. Добавляем подсеть, которая пересекается с whitelist (без force → ошибка)
	ip3 := mustCIDR("10.10.0.0/16")
	s.whitelist[ip3.String()] = &ip3
	_, err = s.AddToBlacklist(ip3, false)
	if err == nil || err.Error() != fmt.Sprintf("IP overlap in whitelist: %s", ip3.String()) {
		t.Errorf("expected whitelist overlap error, got %v", err)
	}

	// 5. С force → должно пройти
	ok, err = s.AddToBlacklist(ip3, true)
	if !ok || err != nil {
		t.Errorf("expected force insert success, got ok=%v err=%v", ok, err)
	}
}

func TestPartialOverlap(t *testing.T) {
	w := NewInMemoryStorage()
	ip1 := mustCIDR("192.168.1.0/25")
	ok, err := w.AddToWhitelist(ip1, false)
	if !ok || err != nil {
		t.Fatalf("expected success adding first subnet, got ok=%v err=%v", ok, err)
	}
	ip2 := mustCIDR("192.168.1.64/25")
	ok, err = w.AddToWhitelist(ip2, false)
	if ok || err == nil {
		t.Errorf("expected partial overlap warning, got ok=%v err=%v", ok, err)
	}
	if _, found := w.whitelist[ip2.String()]; !found {
		t.Errorf("partial overlapping subnet should be added to whitelist")
	}
	b := NewInMemoryStorage()
	ip3 := mustCIDR("10.0.0.0/25")
	ok, err = b.AddToBlacklist(ip3, false)
	if !ok || err != nil {
		t.Fatalf("expected success adding first subnet, got ok=%v err=%v", ok, err)
	}
	ip4 := mustCIDR("10.0.0.64/25")
	ok, err = b.AddToBlacklist(ip4, false)
	if ok || err == nil {
		t.Errorf("expected partial overlap warning, got ok=%v err=%v", ok, err)
	}
	if _, found := b.blacklist[ip4.String()]; !found {
		t.Errorf("partial overlapping subnet should be added to blacklist")
	}
}
