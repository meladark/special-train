package netutils

import (
	"encoding/binary"
	"net"
)

func networkRange(n *net.IPNet) (uint32, uint32) {
	ip4 := n.IP.To4()
	if ip4 == nil {
		panic("only IPv4 supported")
	}
	mask := n.Mask
	if len(mask) == 16 {
		mask = mask[12:]
	}
	if len(mask) != 4 {
		panic("mask must be convertible to 4 bytes for IPv4")
	}
	ipInt := binary.BigEndian.Uint32(ip4)
	maskInt := binary.BigEndian.Uint32(mask)
	network := ipInt & maskInt
	broadcast := network | (^maskInt)
	return network, broadcast
}

func Overlaps(a, b *net.IPNet) (bool, *net.IPNet) {
	aMin, aMax := networkRange(a)
	bMin, bMax := networkRange(b)
	aContainsB := bMin <= aMin && bMax >= aMax
	bContainsA := aMin <= bMin && aMax >= bMax
	if aContainsB {
		return true, b
	}
	if bContainsA {
		return true, a
	}
	return false, nil
}
