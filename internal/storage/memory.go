package storage

import (
	"fmt"
	"net"
	"sync"

	"github.com/meladark/special-train/pkg/netutils"
	_ "github.com/meladark/special-train/pkg/netutils"
)

type InMemoryStorage struct {
	whitelist map[string]*net.IPNet
	blacklist map[string]*net.IPNet
	mu        sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		whitelist: make(map[string]*net.IPNet),
		blacklist: make(map[string]*net.IPNet),
	}
}

func (s *InMemoryStorage) InWhitelist(ip net.IP) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, n := range s.whitelist {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func (s *InMemoryStorage) InBlacklist(ip net.IP) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, n := range s.blacklist {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func (s *InMemoryStorage) AddToWhitelist(ip net.IPNet, force bool) (Ok bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err = nil
	for _, n := range s.whitelist {
		if netutils.Overlaps(n, &ip) {
			return false, fmt.Errorf("IP overlaps in whitelist: %s", n.String())
		}
	}
	for _, n := range s.blacklist {
		if netutils.Overlaps(n, &ip) {
			if !force {
				return false, fmt.Errorf("IP overlap in blacklist: %s", n.String())
			}
		}
	}
	Ok = true
	s.whitelist[ip.String()] = &ip
	return
}

func (s *InMemoryStorage) AddToBlacklist(ip net.IPNet, force bool) (Ok bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	err = nil
	for _, n := range s.blacklist {
		if netutils.Overlaps(n, &ip) {
			return false, fmt.Errorf("IP overlaps in blacklist: %s", n.String())
		}
	}
	for _, n := range s.whitelist {
		if netutils.Overlaps(n, &ip) {
			if !force {
				return false, fmt.Errorf("IP overlap in whitelist: %s", n.String())
			}
		}
	}
	Ok = true
	s.blacklist[ip.String()] = &ip
	return
}
