package netaddr

import "testing"
import "fmt"

func ExampleParseIPv4Net() {
	net, _ := ParseIPv4Net("10.0.0.0/24")
	fmt.Println(net)
	// Output: 10.0.0.0/24
}

func ExampleNewIPv4Net() {
	ip, _ := ParseIPv4("10.0.0.0")
	m32, _ := NewMask32(24)
	net, _ := NewIPv4Net(ip, m32)
	fmt.Println(net)
	// Output: 10.0.0.0/24
}

func ExampleIPv4Net_Extended() {
	net, _ := ParseIPv4Net("10.0.0.0/24")
	fmt.Println(net.Extended())
	// Output: 10.0.0.0 255.255.255.0
}

func ExampleIPv4Net_Fill() {
	net, _ := ParseIPv4Net("10.0.0.0/24")
	subs,_ := NewIPv4NetList([]string{"10.0.0.0/26"})
	subs = net.Fill(subs)           // fills in the missing subnets
	fmt.Println(subs)
	// Output: [10.0.0.0/26 10.0.0.64/26 10.0.0.128/25]
}

func ExampleIPv4Net_Next() {
	net, _ := ParseIPv4Net("10.0.0.4/30")
	next := net.Next()
	fmt.Println(next)
	// Output: 10.0.0.8/29
}

func ExampleIPv4Net_NextSib() {
	net, _ := ParseIPv4Net("10.0.0.4/30")
	next := net.NextSib()
	fmt.Println(next)
	// Output: 10.0.0.8/30
}

func ExampleIPv4Net_Nth() {
	net, _ := ParseIPv4Net("10.0.0.0/24")
	last := net.Len() - 1
	bcast := net.Nth(last) // the broadcast address
	fmt.Println(bcast)
	// Output: 10.0.0.255
}

func ExampleIPv4Net_NthSubnet() {
	net, _ := ParseIPv4Net("10.0.0.0/24")
	fmt.Println(net.NthSubnet(30,2)) // the 3rd /30 subnet
	// Output: 10.0.0.8/30
}

func ExampleIPv4Net_Prev() {
	net, _ := ParseIPv4Net("10.0.0.8/30")
	prev := net.Prev()
	fmt.Println(prev)
	// Output: 10.0.0.0/29
}

func ExampleIPv4Net_PrevSib() {
	net, _ := ParseIPv4Net("10.0.0.8/30")
	prev := net.PrevSib()
	fmt.Println(prev)
	// Output: 10.0.0.4/30
}

func ExampleIPv4Net_Resize() {
	net, _ := ParseIPv4Net("10.0.0.8/30")
	resized := net.Resize(29)
	fmt.Println(resized)
	// Output: 10.0.0.8/29
}

func ExampleIPv4Net_Summ() {
	net1, _ := ParseIPv4Net("10.0.0.0/30")
	net2, _ := ParseIPv4Net("10.0.0.4/30")
	summd := net1.Summ(net2)
	fmt.Println(summd)
	// Output: 10.0.0.0/29
}

func Test_ParseIPv4Net(t *testing.T) {
	cases := []struct {
		given     string
		prefix    uint
		expectErr bool
	}{
		{" 0.0.0.1 ", 32, false},
		{"0.0.0.0/0", 0, false},
		{"192.168.1.1 255.255.255.0", 24, false},
		{"128.0.0.1  255.0.0.0", 8, false},
		{"10.0.0.1/8/24", 0, true},
		{"10.0.0.0/8 255.0.0.0", 0, true},
	}

	for _, c := range cases {
		_, err := ParseIPv4Net(c.given)
		if err != nil {
			if !c.expectErr {
				t.Errorf("ParseIPv4Net(%s) unexpected parse error: %s", c.given, err.Error())
			}
			continue
		}

		if c.expectErr {
			t.Errorf("ParseIPv4Net(%s) expected error but none raised", c.given)
			continue
		}
	}
}

func Test_NewIPv4Net(t *testing.T) {
	cases := []struct {
		ip        string
		m32       uint
		defMask   string
		expectErr bool
	}{
		{"192.168.1.0", 24, "/32", false},
		{"10.0.0.0", 8, "/32", false},
	}

	for _, c := range cases {
		ip, _ := ParseIPv4(c.ip)
		m32 := initMask32(c.m32)

		// test with m32 provided
		net, err := NewIPv4Net(ip, m32)
		if err != nil {
			if !c.expectErr {
				t.Errorf("NewIPv4Net(%s) unexpected error: %s", ip, err.Error())
			}
			continue
		}

		expected := ip.String() + m32.String()
		if net.String() != expected {
			t.Errorf("NewIPv4Net(%s,%s) Expect: %s  Result: %s", ip, m32, expected, net)
		}

		// test with no m32 provided
		net, err = NewIPv4Net(ip, nil)
		if err != nil {
			if !c.expectErr {
				t.Errorf("Unexpected error: %s", err.Error())
			}
			continue
		}

		expected = ip.String() + c.defMask
		if net.String() != expected {
			t.Errorf("NewIPv4Net(%s) Expect: %s  Result: %s", ip, expected, net)
		}
	}
}

func Test_IPv4Net_Cmp(t *testing.T) {
	cases := []struct {
		ip1 string
		ip2 string
		res int
	}{
		{"1.1.1.0/24", "1.1.2.0/24", -1}, // numerically less
		{"1.1.1.0/24", "1.1.0.0/24", 1},  // numerically greater
		{"1.1.1.0/25", "1.1.1.0/24", -1}, // numerically eq, mask less
		{"1.1.1.0/24", "1.1.1.0/25", 1},  // numerically eq, mask greater
		{"1.1.1.0/24", "1.1.1.0/24", 0},  // eq
	}

	for _, c := range cases {
		ip1, _ := ParseIPv4Net(c.ip1)
		ip2, _ := ParseIPv4Net(c.ip2)

		if res, _ := ip1.Cmp(ip2); res != c.res {
			t.Errorf("%s.Cmp(%s) Expect: %d  Result: %d", ip1, ip2, c.res, res)
		}
	}
}

func Test_IPv4Net_Contains(t *testing.T) {
	cases := []struct {
		net    string
		ip     string
		contains bool
	}{
		{"1.0.0.8/29", "1.0.0.15", true},
		{"1.0.0.8/29", "1.0.0.16", false},
		{"1.0.0.8/29", "1.0.0.7", false},
	}

	for _, c := range cases {
		net,_ := ParseIPv4Net(c.net)
		ip,_ := ParseIPv4(c.ip)
		if c.contains != net.Contains(ip) {
			t.Errorf("%s.Contains(%s) Expect: %v  Result: %v", c.net,c.ip,c.contains,!c.contains)
		}
	}
}

func Test_IPv4Net_Fill(t *testing.T) {
	cases := []struct {
		net    string
		subs   []string
		filled []string
	}{
		{ // filter supernet. remove subnets of subnets. basic fwd fill.
			"10.0.0.0/24",
			[]string{"10.0.0.0/24", "10.0.0.0/8", "10.0.0.8/30", "10.0.0.16/30", "10.0.0.16/29", "10.0.0.24/29"},
			[]string{"10.0.0.0/29", "10.0.0.8/30", "10.0.0.12/30", "10.0.0.16/29", "10.0.0.24/29", "10.0.0.32/27", "10.0.0.64/26", "10.0.0.128/25"},
		},
		{ // basic backfill
			"128.0.0.0/1",
			[]string{"192.0.0.0/2"},
			[]string{"128.0.0.0/2", "192.0.0.0/2"},
		},
		{ // basic fwd fill with non-contiguous subnets
			"1.0.0.0/25",
			[]string{"1.0.0.0/30", "1.0.0.64/26"},
			[]string{"1.0.0.0/30", "1.0.0.4/30", "1.0.0.8/29", "1.0.0.16/28", "1.0.0.32/27", "1.0.0.64/26"},
		},
    { // basic backfill. complex fwd fill that uses 'shrink' of the proposed 1.0.16.0/21 subnet
      "1.0.0.0/19",
      []string{"1.0.8.0/21", "1.0.20.0/24"},
      []string{"1.0.0.0/21","1.0.8.0/21","1.0.16.0/22","1.0.20.0/24","1.0.21.0/24","1.0.22.0/23","1.0.24.0/21"},
    },
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
		list, _ := NewIPv4NetList(c.subs)
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

func Test_IPv4Net_Len(t *testing.T) {
	cases := []struct {
		net string
		n   uint32
	}{
		{"1.0.0.64", 1},
		{"1.0.0.64/26", 64},
		{"1.0.0.0/24", 256},
		{"1.0.0.0/0", 0},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
		if net.Len() != c.n {
			t.Errorf("%s.Len() Expect: %d  Result: %d", net, c.n, net.Len())
		}
	}
}

func Test_IPv4Net_Next(t *testing.T) {
	cases := []struct {
		net  string
		next string
		end  bool
	}{
		{"1.0.0.0/31", "1.0.0.2/31", false},
		{"1.0.0.4/30", "1.0.0.8/29", false},
		{"1.0.0.8/29", "1.0.0.16/28", false},
		{"1.0.0.32/27", "1.0.0.64/26", false},
		{"255.255.255.128/25", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
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

func Test_IPv4Net_NextSib(t *testing.T) {
	cases := []struct {
		net  string
		next string
		end  bool
	}{
		{"255.255.255.0/26", "255.255.255.64/26", false},
		{"255.255.255.64/26", "255.255.255.128/26", false},
		{"255.255.255.128/26", "255.255.255.192/26", false},
		{"255.255.255.192/26", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
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

func Test_IPv4Net_Nth(t *testing.T) {
	cases := []struct {
		given  string
		nth    uint32
		expect string
	}{
		{"192.168.1.0/30", 0, "192.168.1.0"},
		{"192.168.1.0/30", 1, "192.168.1.1"},
		{"192.168.1.0/30", 2, "192.168.1.2"},
		{"192.168.1.0/30", 3, "192.168.1.3"},
		{"192.168.1.0/30", 4, ""},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.given)
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

func Test_IPv4Net_NthSubnet(t *testing.T) {
	cases := []struct {
		given  string
		prefix uint
		nth    uint32
		expect string
	}{
		{"192.168.1.0/24", 30, 0, "192.168.1.0/30"},
		{"192.168.1.0/24", 26, 4, ""},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.given)
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

func Test_IPv4Net_Prev(t *testing.T) {
	cases := []struct {
		net  string
		prev string
		end  bool
	}{
		{"1.0.0.8/30", "1.0.0.0/29", false},
		{"1.0.0.192/26", "1.0.0.128/26", false},
		{"1.0.0.128/26", "1.0.0.0/25", false},
		{"0.0.0.0/26", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
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

func Test_IPv4Net_PrevSib(t *testing.T) {
	cases := []struct {
		net  string
		next string
		end  bool
	}{
		{"0.0.0.192/26", "0.0.0.128/26", false},
		{"0.0.0.128/26", "0.0.0.64/26", false},
		{"0.0.0.64/26", "0.0.0.0/26", false},
		{"0.0.0.0/26", "", true},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
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

func Test_IPv4Net_Rel(t *testing.T) {
	cases := []struct {
		ip1   string
		ip2   string
		isRel bool
		rel   int
	}{
		{"1.1.1.0/24", "1.1.1.0/25", true, 1},    // net eq, supernet
		{"1.1.1.0/25", "1.1.1.0/24", true, -1},   // net eq, subnet
		{"1.1.1.0/24", "1.1.1.0/24", true, 0},    // eq
		{"1.1.1.0/24", "1.1.1.128/25", true, 1},  // net ne, supernet
		{"1.1.1.128/25", "1.1.1.0/24", true, -1}, // net ne, subnet
		{"1.1.1.128/25", "1.1.1.0/25", false, 0}, // unrelated
	}

	for _, c := range cases {
		ip1, _ := ParseIPv4Net(c.ip1)
		ip2, _ := ParseIPv4Net(c.ip2)

		if isRel, rel := ip1.Rel(ip2); isRel != c.isRel || rel != c.rel {
			t.Errorf("%s.Rel(%s) Expect: isRel:%t, rel:%d  Result: isRel:%t, rel:%d",
				ip1, ip2, c.isRel, c.rel, isRel, rel)
		}
	}
}

func Test_IPv4Net_Resize(t *testing.T) {
	cases := []struct {
		net    string
		m      uint
		expect string
	}{
		{"1.0.0.64/26", 24, "1.0.0.0/24"},
		{"1.0.0.0/24", 25, "1.0.0.0/25"},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
		net = net.Resize(c.m)
		if net.String() != c.expect {
			t.Errorf("%s.Resize(%d) Expect: %s  Result: %s", c.net, c.m, c.expect, net)
		}
	}
}

func Test_IPv4Net_String(t *testing.T) {
	cases := []struct {
		given  string
		expect string
	}{
		{"0.0.0.0", "0.0.0.0/32"},
		{"192.168.1.1", "192.168.1.1/32"},
		{"1.2.3.4/26", "1.2.3.0/26"},
		{"10.1.1.1/8", "10.0.0.0/8"},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.given)
		if net.String() != c.expect {
			t.Errorf("%s.String() Expect: %s  Result: %s", c.given, c.expect, net)
		}
	}
}

func Test_IPv4Net_SubnetCount(t *testing.T) {
	cases := []struct {
		net    string
		prefix uint
		expect uint32
	}{
		{"10.0.0.0/24", 25, 2},
		{"10.0.0.0/24", 30, 64},
		{"10.0.0.0/24", 24, 0},
		{"10.0.0.0/24", 33, 0},
		{"0.0.0.0/0", 32, 0},
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
		count := net.SubnetCount(c.prefix)
		if count != c.expect {
			t.Errorf("%s.SubnetCount(%d) Expect: %d  Result: %d", c.net, c.prefix, c.expect, count)
			continue
		}
	}
}

func Test_IPv4Net_Summ(t *testing.T) {
	cases := []struct {
		net    string
		other  string
		expect string
		err    bool
	}{
		{"1.1.1.0/30", "1.1.1.4/30", "1.1.1.0/29", false},  // lesser to greater
		{"1.1.1.16/28", "1.1.1.0/28", "1.1.1.0/27", false}, // greater to lesser
		{"1.1.2.0/30", "1.1.1.4/30", "", true},             // different nets
		{"1.1.1.16/28", "1.1.1.32/28", "", true},           // consecutive but not within bit boundary
		{"1.1.1.0/29", "1.1.1.8/30", "", true},             // within bit boundary, but not same size
	}

	for _, c := range cases {
		net, _ := ParseIPv4Net(c.net)
		other, _ := ParseIPv4Net(c.other)
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

