package storage

import "net"

type Storage interface {
	InWhitelist(ip net.IP) bool
	InBlacklist(ip net.IP) bool
}
