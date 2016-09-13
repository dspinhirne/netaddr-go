package netaddr

import "testing"

func TestParseEUI48(t *testing.T) {
	cases := []struct {
		given     string
		expectErr bool
	}{
		{"aa-bb-cc-dd-ee-ff", false},
		{"aa:bb:cc:dd:ee:ff", false},
		{"aabb.ccdd.eeff", false},
		{"aabbccddeeff", false},
		{"aabbccddeeff00", true},
		{"aa,bb,cc,dd,ee,ff", true},
	}

	for _, c := range cases {
		_, err := ParseEUI48(c.given)
		if err != nil {
			if !c.expectErr {
				t.Errorf("ParseEUI48(%s) unexpected parse error: %s", c.given, err.Error())
			}
		} else if c.expectErr {
			t.Errorf("ParseEUI48(%s) expected error but none raised", c.given)
		}
	}
}

func TestEUI48_Strings(t *testing.T) {
	cases := []struct {
		given  string
		expect string
	}{
		{"aa-bb-cc-dd-ee-ff", "aa-bb-cc-dd-ee-ff"},
		{"aabb.ccdd.eeff", "aa-bb-cc-dd-ee-ff"},
		{"aabbccddeeff", "aa-bb-cc-dd-ee-ff"},
		{"00:50:fe:00:00:01", "00-50-fe-00-00-01"},
	}

	for _, c := range cases {
		eui, _ := ParseEUI48(c.given)
		if eui.String() != c.expect {
			t.Errorf("String() expected %s but was %s", c.expect, eui.String())
		}
	}
}

func TestEUI48_ToEUI64(t *testing.T) {
	cases := []struct {
		given  string
		expect string
	}{
		{"aa-bb-cc-dd-ee-ff", "aa-bb-cc-ff-fe-dd-ee-ff"},
	}

	for _, c := range cases {
		eui, _ := ParseEUI48(c.given)
		eui64 := eui.ToEUI64()
		if eui64.String() != c.expect {
			t.Errorf("%s.ToEUI64() expected %s but was %s", c.given, c.expect, eui64)
		}
	}
}
