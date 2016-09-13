package netaddr

import "testing"
import "fmt"

func ExampleNewIPv6NetList() {
	nets := []string{"1::/64", "2::/64"}
	list, _ := NewIPv6NetList(nets)
	fmt.Println(list)
	// Output: [1::/64 2::/64]
}

func ExampleIPv6NetList_Sort() {
	nets := []string{"1::/64", "2::/64", "1::/16", "::", "::1", "1::/16"}
	list, _ := NewIPv6NetList(nets)
	list.Sort()
	fmt.Println(list)
	// Output: [::/128 ::1/128 1::/64 1::/16 1::/16 2::/64]
}

func ExampleIPv6NetList_Summ() {
	nets := []string{"ff00::/13", "ff08::/14", "ff0c::/14", "ff10::/12", "ff20::/11", "ff40::/10", "ff80::/9"}
	list, _ := NewIPv6NetList(nets)
	list = list.Summ()
	fmt.Println(list)
	// Output: [ff00::/8]
}

func Test_NewIPv6NetList(t *testing.T) {
	cases := []struct {
		given []string
		err   bool
	}{
		{
			[]string{"1::/64", "2::/64", "::"},
			false,
		},
		{
			[]string{"1::/64", "2::/64", "::", "fec0/10"},
			true,
		},
	}

	for _, c := range cases {
		_, err := NewIPv6NetList(c.given)
		if err != nil {
			if !c.err {
				t.Errorf("NewIPv6NetList(%s) unexpected error: %s", c.given, err.Error())
			}
		} else if c.err {
			t.Errorf("NewIPv6NetList(%s) expected error but none raised", c.given)
		}
	}
}

func Test_IPv6NetList_Summ(t *testing.T) {
	cases := []struct {
		given  []string
		expect []string
	}{
		{ // summarize a complete subnet
			[]string{"ff00::/13", "ff08::/14", "ff0c::/14", "ff10::/12", "ff20::/11", "ff40::/10", "ff80::/9"},
			[]string{"ff00::/8"},
		},
		{ // summarize the complete ip space
			[]string{"2::/32", "::1", "fec0::/16", "1::/16", "::/0"},
			[]string{"::/0"},
		},
		{ // mix of out of order, duplicates, non-contiguous, subnet-of
			[]string{"ff80::/9", "ff10::/12", "ff80::/10", "ff20::/12", "fff0::/16", "fff1::/16", "ff80::/10"},
			[]string{"ff10::/12", "ff20::/12", "ff80::/9", "fff0::/15"},
		},
	}

	for _, c := range cases {
		list, err := NewIPv6NetList(c.given)
		if err != nil {
			t.Errorf("%v.Summ() unexpected error: %s", list, err.Error())
		} else {
			list = list.Summ()
			for i, e := range list {
				if e.String() != c.expect[i] {
					t.Errorf("%v.Summ() Expect: %v   Result: %v", c.given, c.expect, list)
					break
				}
			}
		}
	}
}
