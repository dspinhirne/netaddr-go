package netaddr

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseMask128 parses a prefix length string to a Mask128 type.
// Netmask must be in "slash" format (eg. '/64' or just '64').
func ParseMask128(prefixLen string) (*Mask128, error) {
	prefixLen = strings.TrimSpace(prefixLen)
	prefixLen = strings.TrimPrefix(prefixLen, "/")
	u8, err := strconv.ParseUint(prefixLen, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("Error parsing prefix length '%s'. %s", prefixLen, err.Error())
	}
	return NewMask128(uint(u8))
}

// NewMask128 converts an integer, representing the prefix length for an IPv6 network,
// to a Mask128 type. Integer must be from 0 to 128.
func NewMask128(prefixLen uint) (*Mask128, error) {
	if prefixLen > 128 {
		return nil, fmt.Errorf("Netmask length %d is too long for IPv6.", prefixLen)
	}
	return initMask128(prefixLen), nil
}

// Mask128 represents a 128-bit netmask used by IPv6Net.
type Mask128 struct {
	netIdMask  uint64 // mask for the netId portion of the address
	hostIdMask uint64 // mask for the hostId portion of the address
	prefixLen     uint   // prefix length
}

/*
Cmp compares equality with another Mask128. Return:
	* 1 if this Mask128 is larger in capacity
	* 0 if the two are equal
	* -1 if this Mask128 is smaller in capacity
*/
func (m128 *Mask128) Cmp(other *Mask128) int {
	if m128.prefixLen < other.prefixLen {
		return 1
	}
	if m128.prefixLen > other.prefixLen {
		return -1
	}
	return 0
}

// HostIdMask returns the internal uint64 mask for the host portion of the mask.
func (m128 *Mask128) HostIdMask() uint64 {
	return m128.hostIdMask
}

// Len returns the number of IP addresses in this network.
// This is only useful if you have a subnet smaller than a /64 as
// it will always return 0 for prefixes <= 64.
func (m128 *Mask128) Len() uint64 {
	if m128.prefixLen <= 64 {
		return 0
	}
	return m128.hostIdMask ^ F64 + 1 // bit flip the netmask and add 1
}

// NetIdMask returns the internal uint64 mask for the network portion of the mask.
func (m128 *Mask128) NetIdMask() uint64 {
	return m128.netIdMask
}

// PrefixLen returns the prefix length as an Uint.
func (m128 *Mask128) PrefixLen() uint {
	return m128.prefixLen
}

// String returns the prefix length as a string.
func (m128 *Mask128) String() string {
	return fmt.Sprintf("/%d", m128.prefixLen)
}


// NON EXPORTED

// initMask128 creates and inits a Mask32
func initMask128(prefixLen uint) *Mask128 {
	m128 := new(Mask128)
	m128.prefixLen = prefixLen
	if prefixLen <= 64 {
		m128.netIdMask = F64 ^ (F64 >> uint64(prefixLen))
	} else {
		m128.netIdMask = F64
		m128.hostIdMask = F64 ^ (F64 >> uint64(prefixLen-64))
	}
	return m128
}
