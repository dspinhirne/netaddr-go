package netaddr

import "testing"

func Test_ParseIPv6Net(t *testing.T) {
	cases := []struct {
		given     string
		prefix    uint
		expectErr bool
	}{
		{"::", 0, false},
		{" ::1 ", 128, false}, // with leading/trailing whitespace
		{"f:0000:fec0:10::/64", 64, false},
		{"1234:0000:fec0:10:0000:0:0:1/64", 64, false},
		{"1234:0000:fec0:10:0000:0:0:1", 128, false},
		{"1234:0000:fec0:10:0000:0:0:1/", 0, true}, // with slash but no prefix
		{"fec0/10", 0, true},                       // improperly formatted
	}

	for _, c := range cases {
		_, err := ParseIPv6Net(c.given)
		if err != nil {
			if !c.expectErr {
				t.Errorf("ParseIPv6NetNet(%s) unexpected error: %s", c.given, err.Error())
			}
			continue
		}

		if c.expectErr {
			t.Errorf("ParseIPv6NetNet(%s) expected error but none raised", c.given)
			continue
		}
	}
}

func Test_IPv6Net_Cmp(t *testing.T) {
	cases := []struct {
		ip1 string
		ip2 string
		res int
	}{
		{"::/128", "::1/128", -1},  // hostId numerically less
		{"::1/128", "::/128", 1},   // hostId numerically greater
		{"::2/128", "::2/127", -1}, // numerically eq, mask less
		{"::2/127", "::2/128", 1},  // numerically eq, mask greater
		{"::2/128", "::2/128", 0},  // eq
	}

	for _, c := range cases {
		ip1, _ := ParseIPv6Net(c.ip1)
		ip2, _ := ParseIPv6Net(c.ip2)

		if res, _ := ip1.Cmp(ip2); res != c.res {
			t.Errorf("%s.Cmp(%s) Expect: %d  Result: %d", ip1, ip2, c.res, res)
		}
	}
}

func Test_IPv6Net_Contains(t *testing.T) {
	cases := []struct {
		net    string
		ip     string
		contains bool
	}{
		{"1:8::/29", "1:f::", true},
		{"1:8::/29", "1:10::", false},
		{"1:8::/29", "1:7::", false},
	}

	for _, c := range cases {
		net,_ := ParseIPv6Net(c.net)
		ip,_ := ParseIPv6(c.ip)
		if c.contains != net.Contains(ip) {
			t.Errorf("%s.Contains(%s) Expect: %v  Result: %v", c.net,c.ip,c.contains,!c.contains)
		}
	}
}

func Test_IPv6Net_Fill(t *testing.T) {
	cases := []struct {
		net    string
		subs   []string
		filled []string
	}{
		{
			"ff00::/8",
			[]string{"ff08::/14", "fe00::/7", "ff20::/11", "ff20::/12"},
			[]string{"ff00::/13", "ff08::/14", "ff0c::/14", "ff10::/12", "ff20::/11", "ff40::/10", "ff80::/9"},
		},
		{
			"ff00::/121",
			[]string{"ff00::/126", "ff00::/120"},
			[]string{"ff00::/126", "ff00::4/126", "ff00::8/125", "ff00::10/124", "ff00::20/123", "ff00::40/122"},
		},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		list, _ := NewIPv6NetList(c.subs)
		list = net.Fill(list)
		if len(list) != len(c.filled) {
			t.Errorf("%s.Fill(%v) Expected: %v  Result: %v", c.net, c.subs, c.filled, list)
			continue
		}
		for i, e := range c.filled {
			if e != list[i].String() {
				t.Errorf("%s.Fill(%v)  Expected: %v  Result: %v", c.net, c.subs, c.filled, list)
				break
			}
		}
	}
}

func Test_IPv6Net_Len(t *testing.T) {
	cases := []struct {
		net string
		n   uint64
	}{
		{"::1/128", 1},
		{"::1/127", 2},
		{"1::/64", 0},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		if net.Len() != c.n {
			t.Errorf("%s.Len() Expect: %d  Result: %d", net, c.n, net.Len())
		}
	}
}

func Test_IPv6Net_Next(t *testing.T) {
	cases := []struct {
		net  string
		next string
		end  bool
	}{
		{"::/127", "::2/127", false},
		{"::4/126", "::8/125", false},
		{"::1:8000:0:0:0/65", "0:0:0:2::/63", false}, // cross /64 boundary
		{"::2:8000:0:0:0/65", "0:0:0:3::/64", false}, // cross /64 boundary
		{"1::/15", "2::/15", false},
		{"4::/14", "8::/13", false},
		{"ffff::/16", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		net = net.Next()

		if net == nil {
			if !c.end {
				t.Errorf("%s.Next() Expect: %s  Result: nil", c.net, c.next)
			}
			continue
		}

		if net.String() != c.next {
			t.Errorf("%s.Next() Expect: %s  Result: %s", c.net, c.next, net)
		}
	}
}

func Test_IPv6Net_NextSib(t *testing.T) {
	cases := []struct {
		net  string
		next string
		end  bool
	}{
		{"::1:8000:0:0:0/65", "0:0:0:2::/65", false}, // add bits across /64 boundary
		{"0:0:0:1::/64", "0:0:0:2::/64", false},
		{"1::/16", "2::/16", false},
		{"ffff::/16", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		net = net.NextSib()

		if net == nil {
			if !c.end {
				t.Errorf("%s.NextSib() Expect: %s  Result: nil", c.net, c.next)
			}
			continue
		}

		if net.String() != c.next {
			t.Errorf("%s.NextSib() Expect: %s  Result: %s", c.net, c.next, net)
		}
	}
}

func Test_IPv6Net_Nth(t *testing.T) {
	cases := []struct {
		given  string
		nth    uint64
		expect string
	}{
		{"1::0/64", 0, "1::"},
		{"::/127", 0, "::"},
		{"::/127", 1, "::1"},
		{"::/127", 2, ""},
		{"1::/16", 0, ""},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.given)
		nth := net.Nth(c.nth)
		if nth == nil {
			if c.expect != "" {
				t.Errorf("%s.Nth(%d) Expect: %s  Result: nil", c.given, c.nth, c.expect)
			}
		} else if nth.String() != c.expect {
			t.Errorf("%s.Nth(%d) Expect: %s  Result: %s", c.given, c.nth, c.expect, nth)
		}
	}
}

func Test_IPv6Net_NthSubnet(t *testing.T) {
	cases := []struct {
		given  string
		prefix uint
		nth    uint64
		expect string
	}{
		{"1::/24", 30, 0, "1::/30"},
		{"1::", 26, 4, ""},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.given)
		nth := net.NthSubnet(c.prefix,c.nth)
		if nth == nil {
			if c.expect != "" {
				t.Errorf("%s.NthSubnet(%d,%d) Expect: %s  Result: nil", c.given, c.prefix, c.nth, c.expect)
			}
		} else if nth.String() != c.expect {
			t.Errorf("%s.NthSubnet(%d,%d) Expect: %s  Result: %s", c.given, c.prefix, c.nth, c.expect, nth)
		}
	}
}

func Test_IPv6Net_Prev(t *testing.T) {
	cases := []struct {
		net  string
		prev string
		end  bool
	}{
		{"1::8/126", "1::/125", false},
		{"f:0:0:2::/63", "f::/63", false},
		{"f::/63", "e::/16", false},
		{"::", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		net = net.Prev()

		if net == nil {
			if !c.end {
				t.Errorf("%s.Prev() Expect: %s  Result: nil", c.net, c.prev)
			}
			continue
		}

		if net.String() != c.prev {
			t.Errorf("%s.Prev() Expect: %s  Result: %s", c.net, c.prev, net)
		}
	}
}

func Test_IPv6Net_PrevSib(t *testing.T) {
	cases := []struct {
		net  string
		next string
		end  bool
	}{
		{"0:0:0:2::/64", "0:0:0:1::/64", false},
		{"2::/16", "1::/16", false},
		{"::/64", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		net = net.PrevSib()

		if net == nil {
			if !c.end {
				t.Errorf("%s.PrevSib() Expect: %s  Result: nil", c.net, c.next)
			}
			continue
		}

		if net.String() != c.next {
			t.Errorf("%s.PrevSib() Expect: %s  Result: %s", c.net, c.next, net)
		}
	}
}

func Test_IPv6Net_Rel(t *testing.T) {
	cases := []struct {
		ip1   string
		ip2   string
		isRel bool
		rel   int
	}{
		{"1::/63", "1::/64", true, 1},        // net eq, supernet
		{"1::/64", "1::/63", true, -1},       // net eq, subnet
		{"1::/64", "1::/64", true, 0},        // eq
		{"1::/60", "1:0:0:1::/64", true, 1},  // net ne, supernet
		{"1:0:0:1::/64", "1::/60", true, -1}, // net ne, subnet
		{"1::/127", "1::1/128", true, 1},     // net ne, supernet
		{"1::1/128", "1::/127", true, -1},    // net ne, subnet
		{"1::/64", "2::/64", false, 0},       // unrelated
	}

	for _, c := range cases {
		ip1, _ := ParseIPv6Net(c.ip1)
		ip2, _ := ParseIPv6Net(c.ip2)

		if isRel, rel := ip1.Rel(ip2); isRel != c.isRel || rel != c.rel {
			t.Errorf("%s.Rel(%s) Expect: isRel:%t, rel:%d  Result: isRel:%t, rel:%d",
				ip1, ip2, c.isRel, c.rel, isRel, rel)
		}
	}
}

func Test_IPv6Net_Resize(t *testing.T) {
	cases := []struct {
		net    string
		m      uint
		expect string
	}{
		{"1::/63", 64, "1::/64"},
		{"1::/64", 65, "1::/65"},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		net = net.Resize(c.m)
		if net.String() != c.expect {
			t.Errorf("%s.Resize(%d) Expect: %s  Result: %s", c.net, c.m, c.expect, net)
		}
	}
}

func Test_IPv6Net_SubnetCount(t *testing.T) {
	cases := []struct {
		net    string
		prefix uint
		expect uint64
	}{
		{"ff::/8", 9, 2},
		{"ff::/8", 10, 4},
		{"ff::/8", 8, 0},
		{"ff::/8", 129, 0},
		{"ff::/8", 128, 0},
		{"ff::/64", 65, 2},
		{"ff::/64", 66, 4},
		{"ff::/8", 128, 0},
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		count := net.SubnetCount(c.prefix)
		if count != c.expect {
			t.Errorf("%s.SubnetCount(%d) Expect: %d  Result: %d", c.net, c.prefix, c.expect, count)
			continue
		}
	}
}

func Test_IPv6Net_Summ(t *testing.T) {
	cases := []struct {
		net    string
		other  string
		expect string
		err    bool
	}{
		{"1::/128", "1::1/128", "1::/127", false},  // lesser to greater
		{"1::1/128", "1::0/128", "1::/127", false}, // greater to lesser
		{"1::/16", "2::/16", "", true},             // different nets
		{"10::/12", "20::/12", "", true},           // consecutive but not within bit boundary
		{"1::/16", "8::/17", "", true},             // within bit boundary, but not same size
	}

	for _, c := range cases {
		net, _ := ParseIPv6Net(c.net)
		other, _ := ParseIPv6Net(c.other)
		summ := net.Summ(other)

		if summ == nil {
			if !c.err {
				t.Errorf("%s.Summ(%s) unexpected error: could not summarize networks", c.net, c.other)
			}
		} else if summ != nil && c.err {
			t.Errorf("%s.Summ(%s) expected error but none raised.", c.net, c.other)
		} else {
			if summ.String() != c.expect {
				t.Errorf("%s.Summ(%s) Expect: %s  Result: %s", c.net, c.other, c.expect, summ)
				continue
			}
		}
	}
}
