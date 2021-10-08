package netaddr

import "testing"
import "fmt"

func ExampleParseIPv4() {
	ip, _ := ParseIPv4("128.0.0.1")
	fmt.Println(ip)
	// Output: 128.0.0.1
}

func ExampleNewIPv4() {
	ip := NewIPv4(0x80000001)
	fmt.Println(ip)
	// Output: 128.0.0.1
}

func ExampleIPv4Cmp() {
	// how does 10.0.0.0 compare with 10.0.0.1?
	ip0, _ := ParseIPv4("10.0.0.0")
	ip1, _ := ParseIPv4("10.0.0.1")
	fmt.Println(ip0.Cmp(ip1))
	// Output: -1 <nil>
}

func Test_ParseIPv4(t *testing.T) {
	cases := []struct {
		given string
		addr  uint32
		err   bool
	}{
		{" 0.0.0.1 ", 1, false},
		{"0.0.0.0", 0, false},
		{"192.168.1.1", 0xc0a80101, false},
		{"128.128.128.128", 0x80808080, false},
		{"256.0.0.1", 0, true},
		{"a.0.0.1", 0, true},
		{"1. 1.1.1", 0, true},
		{"1", 0, true},
	}

	for _, c := range cases {
		ip, err := ParseIPv4(c.given)
		if err != nil {
			if !c.err {
				t.Errorf("ParseIPv4(%s) unexpected parse error: %s", c.given, err.Error())
			}
			continue
		}

		if c.err {
			t.Errorf("ParseIPv4(%s) expected error but none raised", c.given)
			continue
		}

		if ip.addr != c.addr {
			t.Errorf("ParseIPv4(%s).addr  Expect: %x  Result: %x", c.given, c.addr, ip.addr)
		}
	}
}

func Test_IPv4_Cmp(t *testing.T) {
	cases := []struct {
		ip1 string
		ip2 string
		res int
	}{
		{"1.1.1.0", "1.1.2.0", -1}, // numerically less
		{"1.1.1.0", "1.1.0.0", 1},  // numerically greater
		{"1.1.1.0", "1.1.1.0", 0},  // eq
	}

	for _, c := range cases {
		ip1, _ := ParseIPv4(c.ip1)
		ip2, _ := ParseIPv4(c.ip2)

		if res, _ := ip1.Cmp(ip2); res != c.res {
			t.Errorf("%s.Cmp(%s) Expect: %d  Result: %d", ip1, ip2, c.res, res)
		}
	}
}

func Test_MulticastMac(t *testing.T) {
	cases := []struct {
		ip  string
		mac string
	}{
		{"223.255.255.255", ""},
		{"224.0.0.0", "01-00-5e-00-00-00"},
		{"230.2.3.5", "01-00-5e-02-03-05"},
		{"235.147.18.23", "01-00-5e-13-12-17"},
		{"239.255.255.255", "01-00-5e-7f-ff-ff"},
		{"240.0.0.0", ""},
	}

	for _, c := range cases {
		ip, _ := ParseIPv4(c.ip)
		mac := ip.MulticastMac()
		if mac.String() != c.mac {
			t.Errorf("%s.MulticastMac()  Expect: %s  Result: %s", c.ip, c.mac, mac)
		}
	}
}

func Test_IPv4_String(t *testing.T) {
	cases := []string{"0.0.0.0", "192.168.1.0", "1.2.3.4"}

	for _, c := range cases {
		ip, _ := ParseIPv4(c)
		if ip.String() != c {
			t.Errorf("%s.String() Expect: %s  Result: %s", c, c, ip.String())
		}
	}
}

func Test_Ipv4_ToNet(t *testing.T) {
	ip, _ := ParseIPv4("192.168.1.1")
	net, _ := ParseIPv4Net("192.168.1.1")
	cmp,_ := net.Cmp(ip.ToNet())
	if cmp != 0 {
		t.Errorf("%s.ToNet() Expect: %s  Result: %s", ip, net, ip.ToNet())
	}
}
