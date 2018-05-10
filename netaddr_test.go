package netaddr

import "testing"
import "fmt"

func ExampleIPv4PrefixLen() {
	// what size IPv4 subnet is capable of holding 200 addresses?
	fmt.Println(IPv4PrefixLen(200))
	// Output: 24
}

func ExampleParseIP() {
	net,_ := ParseIP("10.0.0.0")
	fmt.Println(net)
	// Output: 10.0.0.0
}

func ExampleParseIP_6() {
	net,_ := ParseIP("fec0::")
	fmt.Println(net)
	// Output: fec0::
}

func ExampleParseIPNet() {
	net,_ := ParseIPNet("10.0.0.0/24")
	fmt.Println(net)
	// Output: 10.0.0.0/24
}

func ExampleParseIPNet_6() {
	net,_ := ParseIPNet("fec0::/10")
	fmt.Println(net)
	// Output: fec0::/10
}

func Test_IPv4PrefixLen(t *testing.T) {
	cases := []struct {
		given  uint
		expect uint
	}{
		{1, 32},
		{30, 27},
		{254, 24},
		{0xfffe, 16},
		{0xfffffe, 8},
		{0xffffffff, 0},
	}

	for _, c := range cases {
		res := IPv4PrefixLen(c.given)
		if res != c.expect {
			t.Errorf("IPv4PrefixLen(%d) did not yield expected result. %d != %d.", c.given, res, c.expect)
		}
	}
}
