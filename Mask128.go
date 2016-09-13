package netaddr

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseMask128 parses a prefix length string to a Mask128 type.
// Netmask must be in "slash" format (eg. '/64' or just '64').
func ParseMask128(prefix string) (*Mask128, error) {
	prefix = strings.TrimSpace(prefix)
	prefix = strings.TrimPrefix(prefix, "/")
	u8, err := strconv.ParseUint(prefix, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("Error parsing prefix length '%s'. %s", prefix, err.Error())
	}
	return NewMask128(uint(u8))
}

// NewMask128 converts an integer, representing the prefix length for an IPv6 network,
// to a Mask128 type. Integer must be from 0 to 128.
func NewMask128(prefix uint) (*Mask128, error) {
	if prefix > 128 {
		return nil, fmt.Errorf("Netmask length %d is too long for IPv6.", prefix)
	}
	return initMask128(prefix), nil
}

// Mask128 represents a 128-bit netmask used by IPv6Net.
type Mask128 struct {
	netIdMask  uint64 // mask for the netId portion of the address
	hostIdMask uint64 // mask for the hostId portion of the address
	prefix     uint   // prefix length
}

/*
Cmp compares equality with another Mask128. Return:
	* 1 if this Mask128 is larger in capacity
	* 0 if the two are equal
	* -1 if this Mask128 is smaller in capacity
*/
func (m128 *Mask128) Cmp(other *Mask128) int {
	if m128.prefix < other.prefix {
		return 1
	}
	if m128.prefix > other.prefix {
		return -1
	}
	return 0
}

// Len returns the number of IP addresses in this network.
// This is only useful if you have a subnet smaller than a /64 as
// it will always return 0 for prefixes <= 64.
func (m128 *Mask128) Len() uint64 {
	if m128.prefix <= 64 {
		return 0
	}
	return m128.hostIdMask ^ ALL_ONES64 + 1 // bit flip the netmask and add 1
}

// PrefixLen returns the prefix length as an Uint.
func (m128 *Mask128) PrefixLen() uint {
	return m128.prefix
}

// String returns the prefix length as a string.
func (m128 *Mask128) String() string {
	return fmt.Sprintf("/%d", m128.prefix)
}

// UintHost returns the mask for the host id portion of the address as a uint64.
func (m128 *Mask128) UintHost() uint64 {
	return m128.hostIdMask
}

// UintNet returns the mask for the network id portion of the address as a uint64.
func (m128 *Mask128) UintNet() uint64 {
	return m128.netIdMask
}

// NON EXPORTED

// dup creates copy of Mask32
func (m128 *Mask128) dup() *Mask128 {
	return &Mask128{m128.netIdMask, m128.hostIdMask, m128.prefix}
}

// initMask128 creates and inits a Mask32
func initMask128(prefix uint) *Mask128 {
	m128 := new(Mask128)
	m128.prefix = prefix
	if prefix <= 64 {
		m128.netIdMask = ALL_ONES64 ^ (ALL_ONES64 >> uint64(prefix))
	} else {
		m128.netIdMask = ALL_ONES64
		m128.hostIdMask = ALL_ONES64 ^ (ALL_ONES64 >> uint64(prefix-64))
	}
	return m128
}
