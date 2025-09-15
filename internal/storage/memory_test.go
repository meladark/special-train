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

func mustIP(s string) net.IP {
	if ip := net.ParseIP(s); ip != nil {
		return ip
	}
	panic(fmt.Sprintf("invalid IP: %s", s))
}

func TestAddToWhitelist(t *testing.T) { //nolint: dupl
	s := NewInMemoryStorage()
	ip := mustCIDR("192.168.1.0/24")
	ok, err := s.AddToWhitelist(ip, false)
	if !ok || err != nil {
		t.Fatalf("expected success, got ok=%v err=%v", ok, err)
	}
	_, err = s.AddToWhitelist(ip, false)
	if err == nil || err.Error() != fmt.Sprintf("IP already in whitelist: %s", ip.String()) {
		t.Errorf("expected full overlap error, got %v", err)
	}
	ip2 := mustCIDR("192.168.1.128/25")
	ok, err = s.AddToWhitelist(ip2, false)
	if ok || err == nil || err.Error() != fmt.Sprintf("IP already in whitelist: %s", ip.String()) {
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
	ip4 := mustCIDR("8.8.8.1/32")
	ip5 := mustCIDR("8.8.8.2/10")
	s.AddToWhitelist(ip4, true)
	s.AddToWhitelist(ip5, true)
	s.BlackWhiteLists()
	s.RemoveFromWhitelist(ip5)
	s.RemoveFromWhitelist(ip5)
}

func TestAddToBlacklist(t *testing.T) { //nolint: dupl
	s := NewInMemoryStorage()
	ip := mustCIDR("172.16.0.0/16")
	ok, err := s.AddToBlacklist(ip, false)
	if !ok || err != nil {
		t.Fatalf("expected success, got ok=%v err=%v", ok, err)
	}
	_, err = s.AddToBlacklist(ip, false)
	if err == nil || err.Error() != fmt.Sprintf("IP already in blacklist: %s", ip.String()) {
		t.Errorf("expected full overlap error, got %v", err)
	}
	ip2 := mustCIDR("172.16.128.0/17")
	ok, err = s.AddToBlacklist(ip2, false)
	if ok || err == nil || err.Error() != fmt.Sprintf("IP already in blacklist: %s", ip.String()) {
		t.Errorf("expected full overlap (subset), got ok=%v err=%v", ok, err)
	}
	if _, found := s.blacklist[ip2.String()]; found {
		t.Errorf("ip2 should NOT be added to blacklist")
	}
	ip3 := mustCIDR("10.10.0.0/16")
	s.whitelist[ip3.String()] = &ip3
	_, err = s.AddToBlacklist(ip3, false)
	if err == nil || err.Error() != fmt.Sprintf("IP overlap in whitelist: %s", ip3.String()) {
		t.Errorf("expected whitelist overlap error, got %v", err)
	}
	ok, err = s.AddToBlacklist(ip3, true)
	if !ok || err != nil {
		t.Errorf("expected force insert success, got ok=%v err=%v", ok, err)
	}
	ip4 := mustCIDR("8.8.8.1/32")
	ip5 := mustCIDR("8.8.8.2/10")
	s.AddToBlacklist(ip4, true)
	s.AddToBlacklist(ip5, true)
	s.BlackWhiteLists()
	s.RemoveFromBlacklist(ip4)
	s.RemoveFromBlacklist(ip4)
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
	if !w.InWhitelist(mustIP("192.168.1.65")) {
		t.Error("expected IP to be in whitelist")
	}
	if w.InWhitelist(mustIP("192.168.2.128")) {
		t.Error(
			"expected IP to NOT be in whitelist",
		)
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
	if !b.InBlacklist(mustIP("10.0.0.65")) {
		t.Error(
			"expected IP to be in blacklist",
		)
	}
	if b.InBlacklist(mustIP("10.0.1.128")) {
		t.Error(
			"expected IP to NOT be in blacklist",
		)
	}
}
