package netaddr

import "testing"

func Test_ParseMask128(t *testing.T) {
	cases := []struct {
		given      string
		prefixLen     uint
		netIdMask  uint64
		hostIdMask uint64
		expectErr  bool
	}{
		{" 8 ", 8, 0xff00000000000000, 0, false},
		{"/32", 32, 0xffffffff00000000, 0, false},
		{"/128", 128, ALL_ONES64, ALL_ONES64, false},
		{"0", 0, 0, 0, false},
		{"//32", 0, 0, 0, true},
		{"/129", 0, 0, 0, true},
	}

	for _, c := range cases {
		m128, err := ParseMask128(c.given)
		if err != nil {
			if !c.expectErr {
				t.Errorf("ParseMask128(%s) unexpected error: %s: %s", c.given, err.Error())
			}
			continue
		}

		if c.expectErr {
			t.Errorf("ParseMask128(%s) expected error but none raised", c.given)
			continue
		}

		if m128.netIdMask != c.netIdMask || m128.hostIdMask != c.hostIdMask {
			t.Errorf("ParseMask128(%s) mask. Expect: %016x%016x  Result: %016x%016x",
				c.given, m128.netIdMask, m128.hostIdMask, c.netIdMask, c.hostIdMask)
		} else if m128.prefixLen != c.prefixLen {
			t.Errorf("ParseMask128(%s) Expect: %d  Result: %d", c.given, m128.prefixLen, c.prefixLen)
		}
	}
}

func Test_NewMask128(t *testing.T) {
	cases := []struct {
		given      uint
		prefixLen     uint
		netIdMask  uint64
		hostIdMask uint64
		expectErr  bool
	}{
		{8, 8, 0xff00000000000000, 0, false},
		{32, 32, 0xffffffff00000000, 0, false},
		{128, 128, ALL_ONES64, ALL_ONES64, false},
		{0, 0, 0, 0, false},
		{129, 0, 0, 0, true},
	}

	for _, c := range cases {
		m128, err := NewMask128(c.given)
		if err != nil {
			if !c.expectErr {
				t.Errorf("Unexpected error: %s", err.Error())
			}
			continue
		}

		if c.expectErr {
			t.Errorf("Expected error when creating Netmask /%d, but none raised", c.given)
			continue
		}

		if m128.netIdMask != c.netIdMask || m128.hostIdMask != c.hostIdMask {
			t.Errorf("Mask for '%s' did not yield expected result. %016x%016x != %016x%016x",
				c.given, m128.netIdMask, m128.hostIdMask, c.netIdMask, c.hostIdMask)
		} else if m128.prefixLen != c.prefixLen {
			t.Errorf("Mask Length for /%d did not yield expected result. %d != %d", c.given, m128.prefixLen, c.prefixLen)
		}
	}
}

func Test_Mask128_String(t *testing.T) {
	cases := []struct {
		given  uint
		expect string
	}{
		{32, "/32"},
		{128, "/128"},
	}

	for _, c := range cases {
		m128 := initMask128(c.given)
		if m128.String() != c.expect {
			t.Errorf("%d.String() Expect: %s  Result: %s", c.given, m128.String(), c.expect)
		}
	}
}

func Test_Mask128_Cmp(t *testing.T) {
	cases := []struct {
		m1  uint
		m2  uint
		res int
	}{
		{25, 24, -1}, // mask less
		{24, 25, 1},  // mask greater
		{24, 24, 0},  // eq
	}

	for _, c := range cases {
		m1, _ := NewMask128(c.m1)
		m2, _ := NewMask128(c.m2)

		if res := m1.Cmp(m2); res != c.res {
			t.Errorf("%s.Cmp(%s). Expect: %d  Result: %d", m1, m2, res, c.res)
		}
	}
}
