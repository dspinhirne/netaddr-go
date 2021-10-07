package netaddr

import (
	"fmt"
	"strings"
)

// IPv4Net represents an IPv4 network.
type IPv4Net struct {
	base *IPv4
	m32  *Mask32
}

/*
ParseIPv4Net parses a string into an IPv4Net type. Accepts addresses in the form of:
	* single IP (eg. 192.168.1.1 -- defaults to /32)
	* CIDR format (eg. 192.168.1.1/24)
	* extended format (eg. 192.168.1.1 255.255.255.0)
*/
func ParseIPv4Net(addr string) (*IPv4Net, error) {
	addr = strings.TrimSpace(addr)
	var m32 *Mask32

	// parse out netmask
	if strings.Contains(addr, "/") { // cidr format
		addrSplit := strings.Split(addr, "/")
		if len(addrSplit) > 2 {
			return nil, fmt.Errorf("IP address contains multiple '/' characters.")
		}
		addr = addrSplit[0]
		prefixLen := addrSplit[1]
		var err error
		m32, err = ParseMask32(prefixLen)
		if err != nil {
			return nil, err
		}
	} else if strings.Contains(addr, " ") { // extended format
		addrSplit := strings.SplitN(addr, " ", 2)
		addr = addrSplit[0]
		mask := addrSplit[1]
		var err error
		m32, err = ParseMask32(mask)
		if err != nil {
			return nil, err
		}
	}

	// parse ip
	ip, err := ParseIPv4(addr)
	if err != nil {
		return nil, err
	}

	return initIPv4Net(ip, m32), nil
}

// NewIPv4Net creates a IPv4Net type from a IPv4 and Mask32.
// If m32 is nil then default to /32.
func NewIPv4Net(ip *IPv4, m32 *Mask32) (*IPv4Net, error) {
	if ip == nil {
		return nil, fmt.Errorf("Argument ip must not be nil.")
	}
	return initIPv4Net(ip, m32), nil
}

/*
Cmp compares equality with another IPv4Net. Return:
	* 1 if this IPv4Net is numerically greater
	* 0 if the two are equal
	* -1 if this IPv4Net is numerically less

The comparasin is initially performed on using the Cmp() method of the network address,
however, in cases where the network addresses are identical then the netmasks will
be compared with the Cmp() method of the netmask.
*/
func (net *IPv4Net) Cmp(other *IPv4Net) (int, error) {
	if other == nil {
		return 0, fmt.Errorf("Argument other must not be nil.")
	}

	res, err := net.base.Cmp(other.base)
	if err != nil {
		return 0, err
	} else if res != 0 {
		return res, nil
	}

	return net.m32.Cmp(other.m32), nil
}

// Contains returns true if the IPv4Net contains the IPv4
func (net *IPv4Net) Contains(ip *IPv4) bool {
	if ip != nil {
		if net.base.addr == ip.addr & net.m32.mask {
			return true
		}
	}
	return false
}

// Extended returns the network address as a string in extended format.
func (net *IPv4Net) Extended() string {
	return net.base.String() + " " + net.m32.Extended()
}

// Fill returns a copy of the given IPv4NetList, stripped of
// any networks which are not subnets of this IPv4Net, and
// with any missing gaps filled in.
func (net *IPv4Net) Fill(list IPv4NetList) IPv4NetList {
	var subs IPv4NetList
	// get rid of non subnets
	if list != nil && len(list) > 0 {
		for _, e := range list {
			isRel, rel := net.Rel(e)
			if isRel && rel == 1 { // e is a subnet
				subs = append(subs, e)
			}
		}
		// discard subnets of subnets & sort
		subs = subs.discardSubnets().Sort()
	} else {
		return subs
	}

	// fill
	var filled IPv4NetList
	if len(subs) > 0 {
		// bottom fill if base address is missing
		base := net.base.addr
		if subs[0].base.addr != base {
			filled = subs[0].backfill(base)
		}

		// fill gaps between subnets
		for i := 0; i < len(subs); i += 1 {
			sub := subs[i]
			if i+1 < len(subs) {
				filled = append(filled, sub.fwdFill(net,subs[i+1])...)
			} else{
				filled = append(filled, sub.fwdFill(net,nil)...)
			}
		}
	}
	return filled
}

// Len returns the number of IP addresses in this network.
// It will always return 0 for /0 networks.
func (net *IPv4Net) Len() uint32 {
	return net.m32.Len()
}

// Netmask returns the Mask32 used by the IPv4Net.
func (net *IPv4Net) Netmask() *Mask32 {
	return net.m32
}

// Network returns the network address of the IPv4Net.
func (net *IPv4Net) Network() *IPv4 {
	return net.base
}

// Next returns the next largest consecutive IP network
// or nil if the end of the address space is reached.
func (net *IPv4Net) Next() *IPv4Net {
	addr := net.nthNextSib(1)
	if addr == nil { // end of address space reached
		return nil
	}
	return addr.grow()
}

// NextSib returns the network immediately following this one.
// It will return nil if the end of the address space is reached.
func (net *IPv4Net) NextSib() *IPv4Net {
	return net.nthNextSib(1)
}

// Nth returns the IP address at the given index.
// The size of the network may be determined with the Len() method.
// If the range is exceeded then return nil.
func (net *IPv4Net) Nth(index uint32) *IPv4 {
	if index >= net.Len() {
		return nil
	}
	return NewIPv4(net.base.addr + index)
}

// NthSubnet returns the subnet IPv4Net at the given index.
// The number of subnets may be determined with the SubnetCount() method.
// If the range is exceeded  or an invalid prefixLen is provided then return nil.
func (net *IPv4Net) NthSubnet(prefixLen uint, index uint32) *IPv4Net {
	count := net.SubnetCount(prefixLen)
	if count == 0 || index >= count{
		return nil
	}
	sub0 := net.Resize(prefixLen)
	return sub0.nthNextSib(index)
}

// Prev returns the previous largest consecutive IP network
// or nil if the start of the address space is reached.
func (net *IPv4Net) Prev() *IPv4Net {
	resized := net.grow()
	return resized.PrevSib()
}

// PrevSib returns the network immediately preceding this one.
// It will return nil if this is 0.0.0.0.
func (net *IPv4Net) PrevSib() *IPv4Net {
	if net.base.addr == 0 {
		return nil
	}
	shift := 32 - net.m32.prefixLen
	addr := (net.base.addr>>shift - 1) << shift
	return &IPv4Net{NewIPv4(addr), net.m32}
}

/*
Rel determines the relationship to another IPv4Net. The method returns
two values: a bool and an int. If the bool is false, then the two networks
are unrelated and the int will be 0. If the bool is true, then the int will
be interpreted as:
	* 1 if this IPv4Net is the supernet of other
	* 0 if the two are equal
	* -1 if this IPv4Net is a subnet of other
*/
func (net *IPv4Net) Rel(other *IPv4Net) (bool, int) {
	if other == nil {
		return false, 0
	}

	// when networks are equal then we can look exlusively at the netmask
	if net.base.addr == other.base.addr {
		return true, net.m32.Cmp(other.m32)
	}

	// when networks are not equal we can use hostmask to test if they are
	// related and which is the supernet vs the subnet
	netHostmask := net.m32.mask ^ F32
	otherHostmask := other.m32.mask ^ F32
	if net.base.addr|netHostmask == other.base.addr|netHostmask {
		return true, 1
	} else if net.base.addr|otherHostmask == other.base.addr|otherHostmask {
		return true, -1
	}
	return false, 0
}

// Resize returns a copy of the network with an adjusted netmask or nil if an invalid prefixLen is given.
func (net *IPv4Net) Resize(prefixLen uint) *IPv4Net{
	if prefixLen > 32{
		return nil
	}
	m32,_ := NewMask32(prefixLen)
	net,_ = NewIPv4Net(net.base, m32)
	return net
}

// String returns the network address as a string in CIDR format.
func (net *IPv4Net) String() string {
	return net.base.String() + net.m32.String()
}


// SubnetCount returns the number a subnets of a given prefix length that this IPv4Net contains.
// It will return 0 for invalid requests (ie. bad prefix or prefix is shorter than that of this network).
// It will also return 0 if the result exceeds the capacity of uint32 (ie. if you want the # of /32 a /0 will hold)
func (net *IPv4Net) SubnetCount(prefixLen uint) uint32 {
	if prefixLen <= net.m32.prefixLen || prefixLen > 32 {
		return 0
	}
	return 1 << (prefixLen - net.m32.prefixLen)
}

// Summ creates a summary address from this IPv4Net and another or nil if the two networks are incapable of being summarized.
func (net *IPv4Net) Summ(other *IPv4Net) *IPv4Net {
	if other == nil || net.m32.prefixLen != other.m32.prefixLen {
		return nil
	}

	// merge-able networks will be identical if you right shift them
	// by the number of bits in the hostmask + 1.
	shift := 32 - net.m32.prefixLen + 1
	addr := net.base.addr >> shift
	otherAddr := other.base.addr >> shift
	if addr != otherAddr {
		return nil
	}
	return net.Resize(net.m32.prefixLen - 1)
}

func (ip *IPv4Net) Version() uint{return 4}

// NON EXPORTED

// backfill generates subnets between this net and the limit address.
// limit should be < net. will create subnets up to and including limit.
func (net *IPv4Net) backfill(limit uint32) IPv4NetList {
	var nets IPv4NetList
	cur := net
	for {
		prev := cur.Prev()
		if prev == nil || prev.base.addr < limit {
			break
		}
		nets = append(IPv4NetList{prev}, nets...)
		cur = prev
	}
	return nets
}

// fwdFill returns subnets between this net and the limit net. limit should be > net.
func (net *IPv4Net) fwdFill(supernet, limit *IPv4Net) IPv4NetList {
	nets := IPv4NetList{net}
	cur := net
	if limit != nil{ // if limit, then fill gaps between net and limit
		for {
			next := cur.nthNextSib(1)
			// ensure we've not exceed the total address space
			if next == nil{break}
			// ensure we've not exceeded the address space of supernet
			if isRel, _ := supernet.Rel(next); !isRel{break}
			// ensure we've not hit limit
			if next.base.addr == limit.base.addr{break}
			
			// check relationship to limit
			isRel, _ := next.Rel(limit)
			if isRel{ // if isRel, then next must be a supernet of limit. we need to shrink it.
				prefixLen := next.m32.prefixLen
				for{
					prefixLen += 1
					next = &IPv4Net{NewIPv4(next.base.addr), initMask32(prefixLen)}
					if isRel, _ := next.Rel(limit); !isRel{break} // stop when we no longer overlap with limit
				}
			} else{ // otherwise, if unrelated then grow until we hit the limit
				prefixLen := next.m32.prefixLen
				mask := next.m32.mask
				for{
					prefixLen -= 1
					if prefixLen == supernet.m32.prefixLen {break}// break if we've hit the supernet boundary
					mask = mask << 1
					if next.base.addr|mask != mask{break} // break when bit boundary crossed (there are '1' bits in the host portion)
					grown := &IPv4Net{NewIPv4(next.base.addr), initMask32(prefixLen)}
					if isRel, _ := grown.Rel(limit); isRel{break} // if we've overlapped with limit in any way, then break
					next = grown
				}
			}
			nets = append(nets, next)
			cur = next
		}
	} else{ // if no limit, then get next largest sibs until we've exceeded supernet
		for {
			next := cur.Next()
			// ensure we've not exceed the total address space
			if next == nil{break}
			// ensure we've not exceeded the address space of supernet
			if isRel, _ := supernet.Rel(next); !isRel{break}
			nets = append(nets, next)
			cur = next
		}
	}
	return nets
}

// initIPv4Net initializes a new IPv4Net
func initIPv4Net(ip *IPv4, m32 *Mask32) *IPv4Net {
	net := new(IPv4Net)
	if m32 == nil {
		m32 = initMask32(32)
	}
	net.m32 = m32
	net.base = NewIPv4(ip.addr & m32.mask) // set base ip
	return net
}

// grow decreases the prefix length as much as possible without crossing a bit boundary.
func (net *IPv4Net) grow() *IPv4Net {
	addr := net.base.addr
	mask := net.m32.mask
	var prefixLen uint
	for prefixLen = net.m32.prefixLen; prefixLen >= 0; prefixLen -= 1 {
		mask = mask << 1
		if addr|mask != mask || prefixLen == 0 { // bit boundary crossed when there are '1' bits in the host portion
			break
		}
	}
	return &IPv4Net{NewIPv4(addr), initMask32(prefixLen)}
}

// nthNextSib returns the nth next sibling network or nil if address space exceeded.
func (net *IPv4Net) nthNextSib(nth uint32) *IPv4Net {
	shift := 32 - net.m32.prefixLen
	addr := (net.base.addr>>shift + nth) << shift
	if addr == 0 { // we exceeded the address space
		return nil
	}
	return &IPv4Net{NewIPv4(addr), net.m32}
}

