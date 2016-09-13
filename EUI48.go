package netaddr

import (
	"fmt"
	"strconv"
)

/*
Parse an EUI-48 string into an EUI48 type.
This will successfully parse most of the typically used formats such as:
	- aa-bb-cc-dd-ee-ff
	- aa:bb:cc:dd:ee:ff
	- aabb.ccdd.eeff
	- aabbccddeeff

Although, in truth, its not picky about the exact format as long as
it contains exactly 12 hex characters with the optional delimiting characters
'-', ':', or '.'.
*/
func ParseEUI48(eui string) (EUI48, error) {
	eui = cleanupEUI(eui)
	if len(eui) != 12 {
		return 0, fmt.Errorf("Must contain exactly 12 hex characters with optional delimiters.")
	}
	u64, err := strconv.ParseUint(eui, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("Error parsing '%s'. %s", eui, err.Error())
	}
	return EUI48(u64), nil
}

// EUI48 (Extended Unique Identifier 48-bit, or EUI-48) represents a 48-bit hardware address.
// It is typically associated with mac-addresses.
type EUI48 uint64

func (eui EUI48) String() string {
	if eui == 0 {
		return ""
	}
	bites := []byte{
		byte(eui >> 40 & 0xff),
		byte(eui >> 32 & 0xff),
		byte(eui >> 24 & 0xff),
		byte(eui >> 16 & 0xff),
		byte(eui >> 8 & 0xff),
		byte(eui & 0xff),
	}
	return fmt.Sprintf("%02x-%02x-%02x-%02x-%02x-%02x", bites[0], bites[1], bites[2], bites[3], bites[4], bites[5])
}

// ToEUI64 converts this EUI48 into an EUI64 by inserting 0xfffe between the OUI and EUI
func (eui EUI48) ToEUI64() EUI64 {
	eui48 := uint64(eui)
	var eui64 uint64 = (eui48&0xffffff000000)<<16 | (eui48 & 0x000000ffffff) | 0x000000fffe000000
	return EUI64(eui64)
}
