package storage

import "net"

type Storage interface {
	InWhitelist(ip net.IP) bool
	InBlacklist(ip net.IP) bool
	AddToWhitelist(ip net.IPNet, force bool) (bool, error)
	AddToBlacklist(ip net.IPNet, force bool) (bool, error)
}
