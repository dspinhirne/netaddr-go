package netaddr

import (
	"fmt"
	"strings"
)

/*
ParseIPv6Net parses a string into an IPv6Net type. Accepts addresses in the form of:
	* single IP (eg. FE80::1)
	* CIDR format (eg. ::1/128)
*/
func ParseIPv6Net(addr string) (*IPv6Net, error) {
	addr = strings.TrimSpace(addr)
	var m128 *Mask128

	// parse out netmask. default to /128 if none provided
	if strings.Contains(addr, "/") { // cidr format
		addrSplit := strings.Split(addr, "/")
		if len(addrSplit) > 2 {
			return nil, fmt.Errorf("IP address contains multiple '/' characters.")
		}
		addr = addrSplit[0]
		prefixLen := addrSplit[1]
		var err error
		m128, err = ParseMask128(prefixLen)
		if err != nil {
			return nil, err
		}
	}

	// create ip
	ip, err := ParseIPv6(addr)
	if err != nil {
		return nil, err
	}

	return initIPv6Net(ip, m128), nil
}

// NewIPv6Net creates a IPv6Net type from a IPv6 and Mask128.
// If netmask is nil then default to /64 (or /0 for address ::).
func NewIPv6Net(ip *IPv6, m128 *Mask128) (*IPv6Net, error) {
	if ip == nil {
		return nil, fmt.Errorf("Argument ip must not be nil.")
	}

	return initIPv6Net(ip, m128), nil
}

// IPv6Net represents an IPv6 network.
type IPv6Net struct {
	base *IPv6
	m128 *Mask128
}

/*
Cmp compares equality with another IPv6Net. Return:
	* 1 if this IPv6Net is numerically greater
	* 0 if the two are equal
	* -1 if this IPv6Net is numerically less

The comparasin is initially performed on using the Cmp() method of the network address,
however, in cases where the network addresses are identical then the netmasks will
be compared with the Cmp() method of the netmask.
*/
func (net *IPv6Net) Cmp(other *IPv6Net) (int, error) {
	if other == nil {
		return 0, fmt.Errorf("Argument other must not be nil.")
	}

	res, err := net.base.Cmp(other.base)
	if err != nil {
		return 0, err
	} else if res != 0 {
		return res, nil
	}

	return net.m128.Cmp(other.m128), nil
}

// Fill returns a copy of the given IPv6NetList, stripped of
// any networks which are not subnets of this IPv6Net, and
// with any missing gaps filled in.
func (net *IPv6Net) Fill(list IPv6NetList) IPv6NetList {
	var subs IPv6NetList
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
	var filled IPv6NetList
	if len(subs) > 0 {
		// bottom fill if base address is missing
		cmp, _ := net.base.Cmp(subs[0].base)
		if cmp != 0 {
			filled = subs[0].backfill(net.base)
		}

		// fill gaps between subnets
		sib := net.nthNextSib(1)
		var ceil *IPv6
		if sib != nil {
			ceil = sib.base
		} else {
			ceil = NewIPv6(F64, F64)
		}
		for i := 0; i < len(subs); i += 1 {
			sub := subs[i]
			filled = append(filled, sub)
			// we need to define a limit for this round
			var limit *IPv6
			if i+1 < len(subs) {
				limit = subs[i+1].base
			} else {
				limit = ceil
			}
			filled = append(filled, sub.fwdFill(limit)...)
		}
	}
	return filled
}

// Len returns the number of IP addresses in this network.
// This is only useful if you have a subnet smaller than a /64 as
// it will always return 0 for prefixes <= 64.
func (net *IPv6Net) Len() uint64 {
	return net.m128.Len()
}

// String returns the network address as a string in long (uncomrpessed) format.
func (net *IPv6Net) Long() string {
	return net.base.Long() + net.m128.String()
}

// Netmask returns the Mask128 used by the IPv6Net.
func (net *IPv6Net) Netmask() *Mask128 {
	return net.m128
}

// Network returns the network address of the IPv6Net.
func (net *IPv6Net) Network() *IPv6 {
	return net.base
}

// Next returns the next largest consecutive IP network
// or nil if the end of the address space is reached.
func (net *IPv6Net) Next() *IPv6Net {
	next := net.nthNextSib(1)
	if next == nil { // passed end of addr space
		return nil
	}
	return next.grow()
}

// NextSib returns the network immediately following this one.
// It will return nil if the end of the address space is reached.
func (net *IPv6Net) NextSib() *IPv6Net {
	return net.nthNextSib(1)
}

// Nth returns the IP address at the given index.
// If the range is exceeded then return nil.
// This only works for /64 and greater; if the prefix length is < 64 then return nil.
// For /64 networks the max index is F64.
// If the prefix length is > 64 then use the Len() method to deterimine the size of the range.
func (net *IPv6Net) Nth(index uint64) *IPv6 {
	if net.m128.prefixLen < 64 || (net.m128.prefixLen > 64 && index >= net.Len()) {
		return nil
	}
	return NewIPv6(net.base.netId, net.base.hostId+index)
}

// NthSubnet returns the subnet IPv6Net at the given index.
// The number of subnets may be determined with the SubnetCount() method.
// If the range is exceeded  or an invalid prefixLen is provided then return nil.
func (net *IPv6Net) NthSubnet(prefixLen uint, index uint64) *IPv6Net {
	count := net.SubnetCount(prefixLen)
	if count == 0 || index >= count{
		return nil
	}
	sub0 := net.Resize(prefixLen)
	return sub0.nthNextSib(index)
}

// Prev returns the previous largest consecutive IP network
// or nil if the start of the address space is reached.
func (net *IPv6Net) Prev() *IPv6Net {
	resized := net.grow()
	return resized.PrevSib()
}

// PrevSib returns the network immediately preceding this one.
// It will return nil if start of the address space is reached.
func (net *IPv6Net) PrevSib() *IPv6Net {
	if net.base.IsZero(){
		return nil
	}
	var netId,hostId uint64
	if net.m128.prefixLen <= 64{ // easy. just working with netId
		shift := 64 - net.m128.prefixLen
		netId = (net.base.netId>>shift - 1) << shift
		hostId = net.base.hostId
	} else{
		shift := 128 - net.m128.prefixLen
		netId = net.base.netId
		hostId = (net.base.hostId>>shift - 1) << shift
	}
	return &IPv6Net{NewIPv6(netId, hostId), net.m128}
}

/*
Rel determines the relationship to another IPv6Net. The method returns
two values: a bool and an int. If the bool is false, then the two networks
are unrelated and the int will be 0. If the bool is true, then the int will
be interpreted as:
	* 1 if this IPv6Net is the supernet of other
	* 0 if the two are equal
	* -1 if this IPv6Net is a subnet of other
*/
func (net *IPv6Net) Rel(other *IPv6Net) (bool, int) {
	cmp, err := net.base.Cmp(other.base)
	if err != nil {
		return false, 0
	}

	// when networks are equal then we can look exlusively at the netmask
	if cmp == 0 {
		return true, net.m128.Cmp(other.m128)
	}

	// when networks are not equal we can use hostmask to test if they are
	// related and which is the supernet vs the subnet
	netHostmask := []uint64{net.m128.netIdMask ^ F64, net.m128.hostIdMask ^ F64}
	otherHostmask := []uint64{other.m128.netIdMask ^ F64, other.m128.hostIdMask ^ F64}
	if net.base.netId|netHostmask[0] == other.base.netId|netHostmask[0] &&
		net.base.hostId|netHostmask[1] == other.base.hostId|netHostmask[1] {
		return true, 1
	} else if net.base.netId|otherHostmask[0] == other.base.netId|otherHostmask[0] &&
		net.base.hostId|otherHostmask[1] == other.base.hostId|otherHostmask[1] {
		return true, -1
	}
	return false, 0
}

// Resize returns a copy of the network with an adjusted netmask.
func (net *IPv6Net) Resize(prefixLen uint) *IPv6Net{
	if prefixLen > 128{
		return nil
	}
	m128,_ := NewMask128(prefixLen)
	net,_ = NewIPv6Net(net.base, m128)
	return net
}

// String returns the network address as a string in zero-compressed format.
func (net *IPv6Net) String() string {
	return net.base.String() + net.m128.String()
}

// SubnetCount returns the number a subnets of a given prefix length that this IPv6Net contains.
// It will return 0 for invalid requests (ie. bad prefix or prefix is shorter than that of this network).
// It will also return 0 if the result exceeds the capacity of uint64 (ie. if you want the # of /128 a /8 will hold)
func (net *IPv6Net) SubnetCount(prefixLen uint) uint64 {
	if prefixLen <= net.m128.prefixLen || prefixLen > 128 {
		return 0
	}
	if prefixLen <= 64 {
		return 1 << (prefixLen - net.m128.prefixLen)
	} else if prefixLen-net.m128.prefixLen >= 64 { // cant exceed 64 bit response
		return 0
	}
	return 1 << (prefixLen - net.m128.prefixLen)
}

// Summ creates a summary address from this IPv6Net and another or nil if the two networks are incapable of being summarized.
func (net *IPv6Net) Summ(other *IPv6Net) *IPv6Net {
	if other == nil || net.m128.prefixLen != other.m128.prefixLen {
		return nil
	}

	// get relevant portion of address
	var addr, otherAddr uint64
	if net.m128.prefixLen <= 64 {
		shift := 64 - net.m128.prefixLen + 1
		addr = net.base.netId >> shift
		otherAddr = other.base.netId >> shift
	} else {
		shift := 128 - net.m128.prefixLen + 1
		addr = net.base.hostId >> shift
		otherAddr = other.base.hostId >> shift
	}

	// merge-able networks will be identical if you right shift them
	// by the number of bits in the hostmask + 1.
	if addr != otherAddr {
		return nil
	}
	return net.Resize(net.m128.prefixLen - 1)
}

// NON EXPORTED

// backfill generates subnets between this net and the limit address.
// limit should be < net. will create subnets up to and including limit.
func (net *IPv6Net) backfill(limit *IPv6) IPv6NetList {
	var nets IPv6NetList
	cur := net
	for {
		prev := cur.Prev()
		if prev == nil {
			break
		}
		cmp, _ := prev.base.Cmp(limit)
		if cmp == -1 {
			break
		}
		nets = append(IPv6NetList{prev}, nets...)
		cur = prev
	}
	return nets
}

// fwdFill returns subnets between this net and the limit address.
// limit should be > net. will create subnets up to limit.
func (net *IPv6Net) fwdFill(limit *IPv6) IPv6NetList {
	var nets IPv6NetList
	cur := net
	for {
		next := cur.Next()
		if next == nil {
			break
		}
		cmp, _ := next.base.Cmp(limit)
		if cmp >= 0 {
			break
		}
		nets = append(nets, next)
		cur = next
	}
	return nets
}

// initIPv6Net initializes a new IPv6Net
func initIPv6Net(ip *IPv6, m128 *Mask128) *IPv6Net {
	net := new(IPv6Net)
	if m128 == nil {
		var prefixLen uint = 64                  // use /64 mask per rfc 4291
		if ip.netId&0x1fffffffffffffff == 0 { // use /128 mask per rfc 4291
			prefixLen = 128
		}
		m128 = initMask128(prefixLen)
	}
	
	// set base ip for this network
	net.base = new(IPv6)
	net.base.netId = ip.netId & m128.netIdMask
	net.base.hostId = ip.hostId & m128.hostIdMask
	net.m128 = m128
	return net
}

// grow decreases the netmask as much as possible without crossing a bit boundary
func (net *IPv6Net) grow() *IPv6Net {
	longPrefix := net.m128.prefixLen > 64 // is the prefix longer than /64
	var prefixLen uint
	var addr, mask uint64
	if longPrefix {
		mask = net.m128.hostIdMask
		addr = net.base.hostId
		prefixLen = net.m128.prefixLen - 64
	} else {
		mask = net.m128.netIdMask
		addr = net.base.netId
		prefixLen = net.m128.prefixLen
	}

	for ; prefixLen >= 0; prefixLen -= 1 {
		mask = mask << 1
		if addr|mask != mask || prefixLen == 0 { // bit boundary crossed when there are '1' bits in the host portion
			break
		}
	}

	if longPrefix { // add back the 64 bits we subtracted above
		prefixLen += 64
	}
	resized := &IPv6Net{NewIPv6(net.base.netId, net.base.hostId), initMask128(prefixLen)}
	if prefixLen == 64 && longPrefix { // we were a longPrefix network and we crossed the /64 boundary. need to keep going
		resized = resized.grow()
	}
	return resized
}

// nthNextSib returns the nth next sibling network or nil if address space exceeded.
func (net *IPv6Net) nthNextSib(nth uint64) *IPv6Net {
	var netId,hostId uint64
	// this can be complex since it is possible to cross the /64 boundary
	if net.m128.prefixLen <= 64{ // easy. just working with netId
		shift := 64 - net.m128.prefixLen
		netId = (net.base.netId>>shift + nth) << shift
		hostId = net.base.hostId
	} else{
		shift := 128 - net.m128.prefixLen
		netId = net.base.netId
		hostId = net.base.hostId >> shift
		if shift < 32 && nth <= uint64(F32){ // easy. just working with hostId since we cant exceed its 64-bit capacity
			hostId = (hostId + nth) << shift
		} else{ // less easy. we need to work with netId and hostId
			hostId = (netId << (64-shift)) | hostId // borrow some bits from netId into the hostId
			netId = netId >> shift
			hostId += nth
			netId = (netId << shift) | (hostId >> (64-shift)) // put our borrowed bits back. but modified
			hostId = hostId<<shift // restore hostId
		}
	}
	ip := NewIPv6(netId, hostId)
	if ip.IsZero(){ // we exceeded the address space
		return nil
	}
	return &IPv6Net{ip, net.m128}
}

