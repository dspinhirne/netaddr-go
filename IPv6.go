package netaddr

import (
	"fmt"
	"strings"
)

type IPv6 struct {
	netId  uint64 // upper 64 bits
	hostId uint64 // lower 64 bits
	str    string // cached String()
}

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
	var ipv4Int uint32
	if strings.Contains(ip, ".") { // check for ipv4 embedded addresses
		elems := strings.Split(ip, ":")
		ipv4,err := ParseIPv4(elems[len(elems)-1])
		if err != nil{
			return nil, fmt.Errorf("Error parsing '%s'. IPv4-embedded IPv6 address is invalid.", ip)
		}
		ip = strings.Replace(ip, elems[len(elems)-1], "0:0", 1) // temporarily remove the ipv4 portion
		ipv4Int = ipv4.addr
	}
	
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
		if numGroups > 8 {
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
	
	// append ipv4-embedded
	if ipv4Int > 0{
		addr.hostId = addr.hostId | uint64(ipv4Int)
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

/*
Cmp compares equality with another IPv6. Return:
	* 1 if this IPv6 is numerically greater than other
	* 0 if the two are equal
	* -1 if this IPv6 is numerically less than other
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

// IPv4 generates an IPv4 from an IPv6 address. The IPv4 address is generated based on the mechanism described by RFC 6052.
// The argument pl (prefix length) should be one of: 32, 40, 48, 56, 64, or 96. Defaults to 96 unless one of the supported values is provided.
func (ip *IPv6) IPv4(pl int) *IPv4{
	if pl == 32{
		return NewIPv4(uint32(ip.netId)) // ipv4 is in lower 32 of net id
	} else if pl == 40{
		i := uint32(ip.hostId >> 48) & 0xff // get the last 8 bits into position
		i2 := (uint32(ip.netId) << 8) & 0xffffff00 // get bottom 24 of net id into position
		return NewIPv4(i | i2)
	} else if pl == 48{
		i := uint32(ip.hostId >> 40) & 0xffff // get the last 16 bits into position
		i2 := (uint32(ip.netId) << 16) & 0xffff0000 // get bottom 16 of net id into position
		return NewIPv4(i | i2)
	} else if pl == 56{
		i := uint32(ip.hostId >> 32) & 0xffffff // get the last 24 bits into position
		i2 := (uint32(ip.netId) << 24) & 0xff000000 // get bottom 8 of net id into position
		return NewIPv4(i | i2)
	} else if pl == 64{
		return NewIPv4(uint32(ip.hostId >> 24)) // get relevant bits of host id into position
	}
	return NewIPv4(uint32(ip.hostId))
}

// IsZero returns true if this address is "::"
func (ip *IPv6) IsZero() bool{
	if ip.netId | ip.hostId == 0{
		return true
	}
	return false
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
	if ip.hostId == F64{
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

	// compress if we've found a series of zero fields in a row.
	// per https://tools.ietf.org/html/rfc5952#section-4.2.2 we must not compress just a single 16-bit zero field.
	if finalEnd-finalStart > 1 {
		head := strings.Join(hexStr[:finalStart], ":")
		tail := strings.Join(hexStr[finalEnd:], ":")
		return head + "::" + tail
	}
	return strings.Join(hexStr, ":")
}

// ToNet returns the IPv6 as a IPv6Net
func (ip *IPv6) ToNet() *IPv6Net{
	return initIPv6Net(ip,nil)
}

func (ip *IPv6) Version() uint{return 6}
