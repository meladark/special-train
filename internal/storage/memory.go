package storage

import (
	"bytes"
	"fmt"
	"net"
	"sync"

	"github.com/meladark/special-train/pkg/netutils"
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

func IPNetEqual(a, b *net.IPNet) bool {
	return a.IP.Equal(b.IP) && bytes.Equal(a.Mask, b.Mask)
}

func (s *InMemoryStorage) addIP(
	targetMap map[string]*net.IPNet,
	otherMap map[string]*net.IPNet,
	ip net.IPNet,
	force bool,
	listName string,
	otherListName string,
) (ok bool, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for oldKey, n := range targetMap {
		if res, ovlp := netutils.Overlaps(n, &ip); res {
			if IPNetEqual(ovlp, n) {
				return false, fmt.Errorf("IP already in %s: %s", listName, n.String())
			}
			delete(targetMap, oldKey)
			targetMap[ip.String()] = ovlp
			return true, fmt.Errorf("IP overlaps in %s: %s", listName, n.String())
		}
	}
	for _, n := range otherMap {
		if res, _ := netutils.Overlaps(n, &ip); res {
			if !force {
				return false, fmt.Errorf("IP overlap in %s: %s", otherListName, n.String())
			}
		}
	}
	targetMap[ip.String()] = &ip
	return true, nil
}

func (s *InMemoryStorage) AddToWhitelist(ip net.IPNet, force bool) (bool, error) {
	return s.addIP(s.whitelist, s.blacklist, ip, force, "whitelist", "blacklist")
}

func (s *InMemoryStorage) AddToBlacklist(ip net.IPNet, force bool) (bool, error) {
	return s.addIP(s.blacklist, s.whitelist, ip, force, "blacklist", "whitelist")
}

func (s *InMemoryStorage) BlackWhiteLists() (whitelist map[string]*net.IPNet, blacklist map[string]*net.IPNet) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.whitelist, s.blacklist
}

func (s *InMemoryStorage) RemoveFromWhitelist(ip net.IPNet) (bool, error) {
	if _, ok := s.whitelist[ip.String()]; ok {
		delete(s.whitelist, ip.String())
		return true, nil
	}
	return false, fmt.Errorf("IP not in whitelist: %s", ip.String())
}

func (s *InMemoryStorage) RemoveFromBlacklist(ip net.IPNet) (bool, error) {
	if _, ok := s.blacklist[ip.String()]; ok {
		delete(s.blacklist, ip.String())
		return true, nil
	}
	return false, fmt.Errorf("IP not in blacklist: %s", ip.String())
}
