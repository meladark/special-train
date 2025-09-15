package netutils

import (
	"net"
	"testing"
)

func mustCIDR(s string) *net.IPNet {
	_, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipnet
}

func TestOverlaps(t *testing.T) {
	tests := []struct {
		a, b   string
		expect bool
	}{
		{"10.0.0.64/26", "10.0.0.128/26", false},
		{"192.168.1.64/26", "192.168.1.128/26", false},
		{"192.168.1.0/25", "192.168.1.0/24", true},   // b больше, чем a
		{"192.168.1.0/24", "192.168.1.128/25", true}, // a включает b
		{"192.168.1.128/25", "192.168.1.0/24", true}, // b включает a
		{"192.168.1.0/24", "192.168.2.0/24", false},  // разные сети
		{"10.0.0.0/8", "10.5.0.0/16", true},          // a включает b
		{"172.16.0.0/16", "172.16.128.0/17", true},   // a включает b
		{"192.168.1.0/25", "192.168.1.64/25", true},  // частичное пересечение
	}

	for _, tt := range tests {
		a := mustCIDR(tt.a)
		b := mustCIDR(tt.b)
		got, _ := Overlaps(a, b)
		if got != tt.expect {
			t.Errorf("Overlaps(%s, %s) = %v, want %v", tt.a, tt.b, got, tt.expect)
		}
	}
}

func TestIsIPv6(t *testing.T) {
	a := mustCIDR("2001:db8::/64")
	b := mustCIDR("2001:db8:0:1000::/65")
	expectedPanic := "only IPv4 supported"
	defer func() {
		if r := recover(); r != nil {
			if r != expectedPanic {
				t.Errorf("ожидалась паника: %q, получено: %q", expectedPanic, r)
			}
		} else {
			t.Error("ожидалась паника, но её не было")
		}
	}()
	Overlaps(a, b)
}
