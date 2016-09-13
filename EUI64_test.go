package netaddr

import "testing"
import "fmt"

func ExampleEUI64_ToIPv6() {
	net, _ := ParseIPv6Net("fe80::/64")
	eui48 := EUI48(0xaabbccddeeff)
	ip := eui48.ToEUI64().ToIPv6(net)
	fmt.Println(ip)
	// Output: fe80::a8bb:ccff:fedd:eeff
}

func TestParseEUI64(t *testing.T) {
	cases := []struct {
		given     string
		expectErr bool
	}{
		{"aa-bb-cc-dd-ee-ff-00-11", false},
		{"aa:bb:cc:dd:ee:ff:00:11", false},
		{"aabb.ccdd.eeff.0011", false},
		{"aabbccddeeff0011", false},
		{"aabbccddeeff001122", true},
		{"aa,bb,cc,dd,ee,ff,00,11", true},
	}

	for _, c := range cases {
		_, err := ParseEUI64(c.given)
		if err != nil {
			if !c.expectErr {
				t.Errorf("ParseEUI64(%s) unexpected parse error: %s", c.given, err.Error())
			}
		} else if c.expectErr {
			t.Errorf("ParseEUI64(%s) expected error but none raised", c.given)
		}
	}
}

func TestEUI64_Strings(t *testing.T) {
	cases := []struct {
		given  string
		expect string
	}{
		{"aa-bb-cc-dd-ee-ff-00-11", "aa-bb-cc-dd-ee-ff-00-11"},
		{"aabb.ccdd.eeff.0011", "aa-bb-cc-dd-ee-ff-00-11"},
		{"aabbccddeeff0011", "aa-bb-cc-dd-ee-ff-00-11"},
		{"aa:bb:cc:dd:ee:ff:00:11", "aa-bb-cc-dd-ee-ff-00-11"},
	}

	for _, c := range cases {
		eui, _ := ParseEUI64(c.given)
		if eui.String() != c.expect {
			t.Errorf("String() expected %s but was %s", c.expect, eui.String())
		}
	}
}

func TestEUI64_ToIPv6(t *testing.T) {
	cases := []struct {
		mac    string
		net    string
		expect string
	}{
		{"aa-bb-cc-dd-ee-ff-00-11", "fe80::/64", "fe80::a8bb:ccdd:eeff:11"},
	}

	for _, c := range cases {
		eui, _ := ParseEUI64(c.mac)
		net, _ := ParseIPv6Net(c.net)
		ip := eui.ToIPv6(net)
		if ip == nil {
			if c.expect != "" {
				t.Errorf("%s.ToIPv6() expected %s but nil received.", c.mac, c.expect)
			}
		} else if ip.String() != c.expect {
			t.Errorf("%s.ToIPv6() expected %s but was %s", c.mac, c.expect, ip)
		}
	}
}
