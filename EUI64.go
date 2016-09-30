package netaddr

import (
	"fmt"
	"strconv"
)

/*
Parse an EUI-64 string into an EUI64 type.
This will successfully parse most of the typically used formats such as:
	- aa-bb-cc-dd-ee-ff-00-11
	- aa:bb:cc:dd:ee:ff:00:11
	- aabb.ccdd.eeff.0011
	- aabbccddeeff0011

Although, in truth, its not picky about the exact format as long as
it contains exactly 16 hex characters with the optional delimiting characters
'-', ':', or '.'.
*/
func ParseEUI64(eui string) (EUI64, error) {
	eui = cleanupEUI(eui)
	if len(eui) != 16 {
		return 0, fmt.Errorf("Must contain exactly 16 characters with optional delimiters.", eui)
	}
	u64, err := strconv.ParseUint(eui, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("Error parsing '%s'. %s", eui, err.Error())
	}
	return EUI64(u64), nil
}

// EUI64 (Extended Unique Identifier 64-bit, or EUI-64) represents a 64-bit hardware address.
type EUI64 uint64

// Bytes returns a slice containing each byte of the EUI64. 
func (eui EUI64) Bytes() []byte {
	return []byte{
		byte(eui >> 56 & 0xff),
		byte(eui >> 48 & 0xff),
		byte(eui >> 40 & 0xff),
		byte(eui >> 32 & 0xff),
		byte(eui >> 24 & 0xff),
		byte(eui >> 16 & 0xff),
		byte(eui >> 8 & 0xff),
		byte(eui & 0xff),
	}
}

func (eui EUI64) String() string {
	if eui == 0 {
		return ""
	}
	bites := eui.Bytes()
	return fmt.Sprintf("%02x-%02x-%02x-%02x-%02x-%02x-%02x-%02x", bites[0], bites[1], bites[2], bites[3],
		bites[4], bites[5], bites[6], bites[7])
}

// ToIPv6 generates an IPv6 address from this EUI64 address and the provided IPv6Net.
// Nil will be returned if IPv6Net is not a /64.
func (eui EUI64) ToIPv6(net *IPv6Net) *IPv6 {
	if net.m128.prefixLen != 64 {
		return nil
	}

	// set u/l bit to 0
	hostId := uint64(eui) ^ 0x0200000000000000
	return NewIPv6(net.base.netId, hostId)
}
