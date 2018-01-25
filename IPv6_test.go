package netaddr

import "testing"

func Test_ParseIPv6(t *testing.T) {
	cases := []struct {
		given     string
		hi64      uint64
		lo64      uint64
		expectErr bool
	}{
		{" :: ", 0, 0, false},
		{"::0", 0, 0, false},
		{"fe80::", 0xfe80000000000000, 0, false},
		//{"::ffff:192.168.1.1",0,0xffffc0a80101,false}, // ipv4 mapped
		{"fe80::1::", 0, 0, true},
		{"::fe80::", 0, 0, true},
		{"0:0:0:0:0:0:0:0:1", 0, 0, true},
		{"::0:0:0:0:0:0:1", 0, 1, false},
		{"1:2:3:4:5:6:7:8", 0x0001000200030004, 0x0005000600070008, false},
		{"1:2:3:4:5:6:7::", 0x0001000200030004, 0x0005000600070000, false},
		{"1:2:3:4:5:6::",   0x0001000200030004, 0x0005000600000000, false},
		{"1:2:3:4:5::",     0x0001000200030004, 0x0005000000000000, false},
		{"1:2:3:4::",       0x0001000200030004, 0x0000000000000000, false},
		{"1:2:3::",         0x0001000200030000, 0x0000000000000000, false},
		{"1:2::",           0x0001000200000000, 0x0000000000000000, false},
		{"1::",             0x0001000000000000, 0x0000000000000000, false},
		{"::1",             0x0000000000000000, 0x0000000000000001, false},
		{"::1:2",           0x0000000000000000, 0x0000000000010002, false},
		{"::1:2:3",         0x0000000000000000, 0x0000000100020003, false},
		{"::1:2:3:4",       0x0000000000000000, 0x0001000200030004, false},
		{"::1:2:3:4:5",     0x0000000000000001, 0x0002000300040005, false},
		{"::1:2:3:4:5:6",   0x0000000000010002, 0x0003000400050006, false},
		{"::1:2:3:4:5:6:7", 0x0000000100020003, 0x0004000500060007, false},
		{"fec0", 0, 0, true},
		{"fec0:::1", 0, 0, true},
	}

	for _, c := range cases {
		ip, err := ParseIPv6(c.given)
		if err != nil {
			if !c.expectErr {
				t.Errorf("ParseIPv6(%s) unexpected parse error: %s", c.given, err.Error())
			}
			continue
		}

		if c.expectErr {
			t.Errorf("ParseIPv6(%s) expected error but none raised", c.given)
			continue
		}

		if ip.netId != c.hi64 || ip.hostId != c.lo64 {
			t.Errorf("ParseIPv6(%s)  Expect: %x%x  Result: %x%x", c.given, c.hi64, c.lo64, ip.netId, ip.hostId)
		}
	}
}

func Test_IPv6_Cmp(t *testing.T) {
	cases := []struct {
		ip1 string
		ip2 string
		res int
	}{
		{"::", "::1", -1},  // hostId numerically less
		{"::1", "::", 1},   // hostId numerically greater
		{"::1", "::1", 0},  // hostId eq
		{"1::", "2::", -1}, // netId numerically less
		{"2::", "1::", 1},  // netId numerically greater
		{"1::", "1::", 0},  // netId eq
	}

	for _, c := range cases {
		ip1, _ := ParseIPv6(c.ip1)
		ip2, _ := ParseIPv6(c.ip2)

		if res, _ := ip1.Cmp(ip2); res != c.res {
			t.Errorf("%s.Cmp(%s) Expect: %d  Result: %d", ip1, ip2, c.res, res)
		}
	}
}

func Test_IPv6_Long(t *testing.T) {
	cases := []struct {
		given  string
		expect string
	}{
		{"::", "0000:0000:0000:0000:0000:0000:0000:0000"},
		{"1::", "0001:0000:0000:0000:0000:0000:0000:0000"},
		{"1000::", "1000:0000:0000:0000:0000:0000:0000:0000"},
	}

	for _, c := range cases {
		ip, _ := ParseIPv6(c.given)
		long := ip.Long()
		if long != c.expect {
			t.Errorf("%s.Long() Expect: %s  Result: %s", c.given, c.expect, long)
		}
	}
}

func Test_IPv6_String(t *testing.T) {
	cases := []struct {
		given  string
		expect string
	}{
		{"0:0:0:0:0:0:0:0", "::"},
		{"1:0:0:0:0:0:0:0", "1::"},
		{"0:1:0:0:0:0:0:0", "0:1::"},
		{"0:0:1:0:0:0:0:0", "0:0:1::"},
		{"0:0:0:1:0:0:0:0", "0:0:0:1::"},
		{"0:0:0:0:1:0:0:0", "::1:0:0:0"},
		{"0:0:0:0:0:1:0:0", "::1:0:0"},
		{"0:0:0:0:0:0:1:0", "::1:0"},
		{"0:0:0:0:0:0:0:1", "::1"},
		
		{"1:0:0:0:0:0:0:1", "1::1"},
		{"1:1:0:0:0:0:0:1", "1:1::1"},
		{"1:0:1:0:0:0:0:1", "1:0:1::1"},
		{"1:0:0:1:0:0:0:1", "1:0:0:1::1"},
		{"1:0:0:0:1:0:0:1", "1::1:0:0:1"},
		{"1:0:0:0:0:1:0:1", "1::1:0:1"},
		{"1:0:0:0:0:0:1:1", "1::1:1"},
	}

	for _, c := range cases {
		ip, _ := ParseIPv6(c.given)
		short := ip.String()
		if short != c.expect {
			t.Errorf("%s.String() Expect: %s  Result: %s", c.given, c.expect, short)
		}
	}
}

func Test_Ipv6_ToNet(t *testing.T) {
	ip, _ := ParseIPv6("1::")
	net, _ := ParseIPv6Net("1::")
	cmp,_ := net.Cmp(ip.ToNet())
	if cmp != 0 {
		t.Errorf("%s.ToNet() Expect: %s  Result: %s", ip, net, ip.ToNet())
	}
}
