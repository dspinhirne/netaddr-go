/*
netaddr is a Go library for performing calculations on IPv4 and IPv6 subnets. There is also limited support for EUI addresses.
*/
package netaddr

import (
	"strconv"
	"strings"
)

const (
	// 32 bits worth of '1'
	F32 uint32 = 0xffffffff

	// 64 bits worth of '1'
	F64 uint64 = 0xffffffffffffffff
)

// IPv4PrefixLen returns the prefix length needed to hold the
// number of IP addresses specified by "size".
func IPv4PrefixLen(size uint) uint {
	var prefix uint
	for prefix = 32; prefix >= 0; prefix -= 1 {
		hostbits := 32 - prefix
		var max uint = 1 << hostbits
		if size <= max {
			break
		}
	}
	return prefix
}

// NON EXPORTED

// cleanupEUI removes delimiter characters from eui address string
func cleanupEUI(addr string) string {
	addr = strings.TrimSpace(addr)
	addr = strings.Replace(addr, ":", "", -1)
	addr = strings.Replace(addr, "-", "", -1)
	addr = strings.Replace(addr, ".", "", -1)
	return addr
}

// u8SlicetoU32 converts a slice of 4 strings representing uint8 numbers (base 10) to a uint32.
func u8SlicetoU32(group []string) (uint32, error) {
	var g uint64 = 4
	var u32 uint32
	for _, e := range group {
		g -= 1
		u8, err := strconv.ParseUint(e, 10, 8)
		if err != nil {
			return 0, err
		}
		u8 = u8 << (8 * g)
		u32 = u32 | uint32(u8)
	}
	return u32, nil
}

// u16SlicetoU64 converts a slice of 4 strings representing uint16 numbers (in hex) to a uint64.
func u16SlicetoU64(group []string) (uint64, error) {
	var g uint64 = 4
	var u64 uint64
	for _, e := range group {
		g -= 1
		u16, err := strconv.ParseUint(e, 16, 16)
		if err != nil {
			return 0, err
		}
		u16 = u16 << (16 * g)
		u64 = u64 | u16
	}
	return u64, nil
}
