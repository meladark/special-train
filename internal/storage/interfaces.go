package storage

import "net"

type Storage interface {
	InWhitelist(ip net.IP) bool
	InBlacklist(ip net.IP) bool
	AddToWhitelist(ip net.IPNet, force bool) (bool, error)
	AddToBlacklist(ip net.IPNet, force bool) (bool, error)
	BlackWhiteLists() (whitelist map[string]*net.IPNet, blacklist map[string]*net.IPNet)
	RemoveFromWhitelist(ip net.IPNet) (bool, error)
	RemoveFromBlacklist(ip net.IPNet) (bool, error)
}
