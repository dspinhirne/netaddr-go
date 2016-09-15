package netaddr

import (
	"fmt"
	"strings"
)

/*
ParseIPv6 arses a string into an IPv6 type.
IP address should be in one of the following formats and should not contain a netmask.
	* long format (eg. 0000:0000:0000:0000:0000:0000:0000:0001)
	* zero-compressed short format (eg. ::1)
*/
func ParseIPv6(ip string) (*IPv6, error) {
	ip = strings.TrimSpace(ip)

	if ip == "::" {
		return new(IPv6), nil
	} // special case. just return zero address

	var groups []string             // holds the 8 groups of hex strings representing the ipv6 addr
	if strings.Contains(ip, "::") { // ip is using shorthand notation
		halves := strings.Split(ip, "::")
		if len(halves) != 2 {
			return nil, fmt.Errorf("Error parsing '%s'. Contains %d '::' sequences.", ip, len(halves))
		}
		if halves[0] == "" {
			halves[0] = "0"
		} // handle cases such as ::1
		if halves[1] == "" {
			halves[1] = "0"
		} // handle cases such as fe80::
		upHalf := strings.Split(halves[0], ":")
		loHalf := strings.Split(halves[1], ":")
		numGroups := len(upHalf) + len(loHalf)
		if numGroups > 6 {
			return nil, fmt.Errorf("Error parsing '%s'. Shorthand formatted address is too long.", ip)
		}
		groups = upHalf
		for i := 8 - numGroups; i > 0; i -= 1 {
			groups = append(groups, "0")
		}
		groups = append(groups, loHalf...)

	} else {
		groups = strings.Split(ip, ":")
		if len(groups) > 8 {
			return nil, fmt.Errorf("Error parsing '%s'. Address is too long.", ip)
		} else if len(groups) < 8 {
			return nil, fmt.Errorf("Error parsing '%s'. Address is too short.", ip)
		}
	}

	addr := new(IPv6)
	if u64, err := u16SlicetoU64(groups[0:4]); err != nil {
		return nil, fmt.Errorf("Error parsing '%s'. %s", ip, err.Error())
	} else {
		addr.netId = u64
	}

	if u64, err := u16SlicetoU64(groups[4:]); err != nil {
		return nil, fmt.Errorf("Error parsing '%s'. %s", ip, err.Error())
	} else {
		addr.hostId = u64
	}

	return addr, nil
}

/*
NewIPv6 creates an IPv6 type from a pair of uint64. The pair represents
the upper/lower 64-bits of the address respectively
*/
func NewIPv6(netId, hostId uint64) *IPv6 {
	return &IPv6{netId: netId, hostId: hostId}
}

type IPv6 struct {
	netId  uint64 // upper 64 bits
	hostId uint64 // lower 64 bits
	str    string // cached String()
}

/*
Cmp compares equality with another IPv6. Return:
	* 1 if this IPv6 is numerically greater
	* 0 if the two are equal
	* -1 if this IPv6 is numerically less
*/
func (ip *IPv6) Cmp(other *IPv6) (int, error) {
	if other == nil {
		return 0, fmt.Errorf("Argument other must not be nil.")
	}

	if ip.netId == other.netId { // compare hostId when netId is eq
		if ip.hostId > other.hostId {
			return 1, nil
		}
		if ip.hostId < other.hostId {
			return -1, nil
		}
	} else if ip.netId > other.netId {
		return 1, nil
	} else if ip.netId < other.netId {
		return -1, nil
	}
	return 0, nil
}

// HostId returns the interal uint64 for the host id portion of the address.
func (ip *IPv6) HostId() uint64 {
	return ip.hostId
}

// Long returns the IPv6 address as a string in long (uncompressed) format.
func (ip *IPv6) Long() string {
	return fmt.Sprintf(
		"%04x:%04x:%04x:%04x:%04x:%04x:%04x:%04x",
		ip.netId>>48&0xffff,
		ip.netId>>32&0xffff,
		ip.netId>>16&0xffff,
		ip.netId&0xffff,
		ip.hostId>>48&0xffff,
		ip.hostId>>32&0xffff,
		ip.hostId>>16&0xffff,
		ip.hostId&0xffff,
	)
}

// NetId returns the interal uint64 for the network id portion of the address.
func (ip *IPv6) NetId() uint64 {
	return ip.netId
}

// Next returns the next consecutive IPv6 or nil if the end of this /64 address space is reached.
func (ip *IPv6) Next() *IPv6 {
	if ip.hostId == ALL_ONES64{
		return nil
	}
	return NewIPv6(ip.netId, ip.hostId + 1)
}

// Prev returns the preceding IPv6 or nil if this is first address of this /64 space.
func (ip *IPv6) Prev() *IPv6 {
	if ip.hostId == 0{
		return nil
	}
	return NewIPv6(ip.netId, ip.hostId - 1)
}

// String returns IPv6 as a string in zero-compressed format (per rfc5952).
// Use Long() to render in uncompressed format.
func (ip *IPv6) String() string {
	hexStr := make([]string, 8, 8)
	u64 := ip.netId
	zeroStart, finalStart, finalEnd, consec0 := -1, -1, -1, 0
	var i uint
	for ; i < 2; i += 1 {
		var i2 uint
		for ; i2 < 4; i2 += 1 {
			// capture 2-byte word
			hexStrI := 4*i + i2
			shift := 48 - 16*i2
			wd := (u64 >> shift) & 0xffff
			hexStr[hexStrI] = fmt.Sprintf("%x", wd)

			// captured count of consecutive zeros
			if wd == 0 {
				if zeroStart == -1 {
					zeroStart = int(hexStrI)
				}
				consec0 += 1
			}

			// test for longest consecutive zeros when non-zero encountered or we're at the end
			if wd != 0 || hexStrI == 7 {
				if consec0 > finalEnd-finalStart {
					finalStart = zeroStart
					finalEnd = finalStart + consec0
				}
				zeroStart = -1
				consec0 = 0
			}
		}
		u64 = ip.hostId
	}

	// compress if we've found a series of 0 words in a row
	if finalStart != -1 {
		head := strings.Join(hexStr[:finalStart], ":")
		tail := strings.Join(hexStr[finalEnd:], ":")
		return head + "::" + tail
	}
	return strings.Join(hexStr, ":")
}

