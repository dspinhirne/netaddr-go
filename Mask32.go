package netaddr

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseMask32 parses an IPv4 netmask or prefix length string to a Mask32 type.
// Netmask must be in either dotted-quad format (y.y.y.y) or "slash"
// format (eg. '/32' or just '32').
func ParseMask32(netmask string) (*Mask32, error) {
	netmask = strings.TrimSpace(netmask)

	// parse cidr format
	if !strings.Contains(netmask, ".") {
		netmask = strings.TrimPrefix(netmask, "/")
		u8, err := strconv.ParseUint(netmask, 10, 8)
		if err != nil {
			return nil, fmt.Errorf("Error parsing CIDR netmask '%s'. %s", netmask, err.Error())
		}
		return NewMask32(uint(u8))
	}

	// parse from extended format
	ip, err := ParseIPv4(netmask)
	if err != nil {
		return nil, err
	}
	u32 := ip.addr

	/*
		determine length of netmask by cycling through bit by bit and looking
		for the first '1' bit, tracking the length as we go. we also want to verify
		that the mask is valid (ie. not something like 255.254.255.0). we do this
		by creating a hostmask which covers the '0' bits of the mask. once we have
		separated the net vs host mask we xor them together. the result should be that
		all bits are now '1'. if not then we know we have an invalid netmask.
	*/
	var prefixLen uint = 32
	var hostmask uint32 = 1
	mask := u32
	for i := 32; i > 0; i -= 1 {
		if u32&1 == 1 {
			hostmask = hostmask >> 1
			if mask^hostmask != F32 {
				return nil, fmt.Errorf("Netmask '%s' is invalid. It contains '1' bits in its host portion.", netmask)
			}
			break
		}
		hostmask = (hostmask << 1) | 1
		u32 = u32 >> 1
		prefixLen -= 1
	}

	return initMask32(prefixLen), nil
}

// NewMask32 converts an integer, representing the prefix length for an IPv4 address,
// to a Mask32 type. Integer must be from 0 to 32.
func NewMask32(prefixLen uint) (*Mask32, error) {
	if prefixLen > 32 {
		return nil, fmt.Errorf("Netmask length %d is too long for IPv4.", prefixLen)
	}
	return initMask32(prefixLen), nil
}

// Mask32 represents a 32-bit netmask used by IPv4Net.
type Mask32 struct {
	mask   uint32
	prefixLen uint // prefix length
}

/*
Cmp compares equality with another Mask32. Return:
	* 1 if this Mask32 is larger in capacity
	* 0 if the two are equal
	* -1 if this Mask32 is smaller in capacity
*/
func (m32 *Mask32) Cmp(other *Mask32) int {
	if m32.prefixLen < other.prefixLen {
		return 1
	}
	if m32.prefixLen > other.prefixLen {
		return -1
	}
	return 0
}

// Extended returns the Mask32 as a string in extended format.
func (m32 *Mask32) Extended() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		m32.mask>>24&0xff,
		m32.mask>>16&0xff,
		m32.mask>>8&0xff,
		m32.mask&0xff)
}

// Len returns the number of IP addresses in this network.
// It will always return 0 for /0 networks.
func (m32 *Mask32) Len() uint32 {
	if m32.mask == 0 {
		return 0
	}
	return m32.mask ^ F32 + 1 // bit flip the netmask and add 1
}

// Mask returns the internal uint32 mask.
func (m32 *Mask32) Mask() uint32 {
	return m32.mask
}

// PrefixLen returns the prefix length as an Uint.
func (m32 *Mask32) PrefixLen() uint {
	return m32.prefixLen
}

// String returns the prefix length as a string.
// Use Extended() to return in extended format instead.
func (m32 *Mask32) String() string {
	return fmt.Sprintf("/%d", m32.prefixLen)
}


// NON EXPORTED


// initMask32 creates and inits a Mask32
func initMask32(prefixLen uint) *Mask32 {
	m32 := &Mask32{prefixLen: prefixLen}
	m32.mask = F32 ^ (F32 >> uint32(prefixLen))
	return m32
}
