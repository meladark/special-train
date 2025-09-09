package storage

import (
	"net"
	"sync"
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
