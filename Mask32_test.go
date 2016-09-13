package netaddr

import "testing"
import "fmt"

func ExampleParseMask32() {
	m32, _ := ParseMask32("/32")
	fmt.Println(m32)
	// Output: /32
}

func ExampleNewMask32() {
	m32, _ := NewMask32(32)
	fmt.Println(m32)
	// Output: /32
}

func ExampleMask32_Extended() {
	m32, _ := ParseMask32("/24")
	fmt.Println(m32.Extended())
	// Output: 255.255.255.0
}

func Test_ParseMask32(t *testing.T) {
	cases := []struct {
		given  string
		prefix uint
		mask   uint32
		err    bool
	}{
		{" 255.0.0.0 ", 8, 0xff000000, false},
		{"0.0.0.0", 0, 0, false},
		{"255.255.255.255", 32, 0xffffffff, false},
		{" 8 ", 8, 0xff000000, false},
		{"/32", 32, 0xffffffff, false},
		{"//32", 0, 0, true},
		{"256.0.0.0", 0, 0, true},
		{"255.248.255.0", 0, 0, true},
		{"255", 0, 0, true},
	}

	for _, c := range cases {
		m32, err := ParseMask32(c.given)
		if err != nil {
			if !c.err {
				t.Errorf("ParseMask32(%s) unexpected error: %s", c.given, err.Error())
			}
			continue
		}

		if c.err {
			t.Errorf("ParseMask32(%s) expected error but none raised", c.given)
			continue
		}

		if m32.mask != c.mask {
			t.Errorf("ParseMask32(%s) mask. Expect: %08x  Result: %08x", c.given, m32.mask, c.mask)
		} else if m32.prefix != c.prefix {
			t.Errorf("ParseMask32(%s) prefix. Expect: %d  Result: %d", c.given, m32.prefix, c.prefix)
		}
	}
}

func Test_NewMask32(t *testing.T) {
	cases := []struct {
		given  uint
		prefix uint
		mask   uint32
		err    bool
	}{
		{8, 8, 0xff000000, false},
		{17, 17, 0xffff8000, false},
		{0, 0, 0, false},
		{32, 32, 0xffffffff, false},
		{33, 0, 0, true},
	}

	for _, c := range cases {
		m32, err := NewMask32(c.given)
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error: %s", err.Error())
			}
			continue
		}

		if c.err {
			t.Errorf("Expected error when creating Netmask /%d, but none raised", c.given)
			continue
		}

		if m32.mask != c.mask {
			t.Errorf("NewMask(%d). Expect: %x  Result: %x", c.given, m32.mask, c.mask)
		} else if m32.prefix != c.prefix {
			t.Errorf("NewMask(%d). Expect: %d  Result: %d", c.given, m32.prefix, c.prefix)
		}
	}
}

func Test_Mask32_Extended(t *testing.T) {
	cases := []struct {
		given    uint
		extended string
	}{
		{32, "255.255.255.255"},
		{8, "255.0.0.0"},
	}

	for _, c := range cases {
		m32 := initMask32(c.given)
		ext := m32.Extended()
		if ext != c.extended {
			t.Errorf("%d.Extended(). Expect: %s  Result: %s", c.given, ext, c.extended)
		}
	}
}

func Test_Mask32_Cmp(t *testing.T) {
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
		m1 := initMask32(c.m1)
		m2 := initMask32(c.m2)

		if res := m1.Cmp(m2); res != c.res {
			t.Errorf("%s.Cmp(%s). Expect: %d  Result: %d", m1, m2, res, c.res)
		}
	}
}

func Test_Mask32_String(t *testing.T) {
	cases := []struct {
		given  uint
		expect string
	}{
		{32, "/32"},
		{8, "/8"},
	}

	for _, c := range cases {
		m32 := initMask32(c.given)
		if m32.String() != c.expect {
			t.Errorf("%d.String(). Expect: %s  Result: %s", c.given, m32.String(), c.expect)
		}
	}
}
