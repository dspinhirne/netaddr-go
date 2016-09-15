package netaddr

import (
	"fmt"
	"strings"
)

// ParseIPv4 parses a string into an IPv4 type.
// IP address should be in dotted-quad format (x.x.x.x) and should not contain a netmask.
func ParseIPv4(ip string) (*IPv4, error) {
	ip = strings.TrimSpace(ip)
	bites := strings.SplitN(ip, ".", 4)
	if len(bites) != 4 {
		return nil, fmt.Errorf("Error parsing '%s'. IPv4 address must have exactly 4 octets.", ip)
	}
	addr, err := u8SlicetoU32(bites)
	if err != nil {
		return nil, fmt.Errorf("Error parsing '%s'. %s", ip, err.Error())
	}
	return &IPv4{addr: addr}, nil
}

// NewIPv4 creates an IPv4 type from a uint32
func NewIPv4(addr uint32) *IPv4 {
	return &IPv4{addr: addr}
}

// IPv4 represents an version 4 IP address.
type IPv4 struct {
	addr uint32
}

// Addr returns the internal uint32 address.
func (ip *IPv4) Addr() uint32 {
	return ip.addr
}

/*
Cmp compares equality with another IPv4. Return:
	* 1 if this IPv4 is numerically greater
	* 0 if the two are equal
	* -1 if this IPv4 is numerically less
*/
func (ip *IPv4) Cmp(other *IPv4) (int, error) {
	if other == nil {
		return 0, fmt.Errorf("Argument other must not be nil.")
	}

	if ip.addr > other.addr {
		return 1, nil
	}
	if ip.addr < other.addr {
		return -1, nil
	}
	return 0, nil
}

// MulticastMac returns the multicast mac-address for this IP.
// It will return a value of 0 for addresses outside of the
// multicast range 224.0.0.0/4.
func (ip *IPv4) MulticastMac() EUI48 {
	var mac EUI48
	if ip.addr&0xf0000000 == 0xe0000000 { // within 224.0.0.0/4 ?
		// map lower 23-bits of ip to 01:00:5e:00:00:00
		mac = EUI48(ip.addr&0x007fffff) | 0x01005e000000
	}
	return mac
}

// Next returns the next consecutive IPv4 or nil if the end of the address space is reached.
func (ip *IPv4) Next() *IPv4 {
	if ip.addr == ALL_ONES32{
		return nil
	}
	return NewIPv4(ip.addr + 1)
}

// Prev returns the preceding IPv4 or nil if this is 0.0.0.0.
func (ip *IPv4) Prev() *IPv4 {
	if ip.addr == 0{
		return nil
	}
	return NewIPv4(ip.addr - 1)
}

// String return IPv4 address as a string.
func (ip *IPv4) String() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ip.addr>>24&0xff,
		ip.addr>>16&0xff,
		ip.addr>>8&0xff,
		ip.addr&0xff)
}
