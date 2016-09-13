package netaddr

import "testing"
import "fmt"

func ExampleNewIPv4NetList() {
	nets := []string{"10.0.0.0/24", "1.0.0.0/24"}
	list, _ := NewIPv4NetList(nets)
	fmt.Println(list)
	// Output: [10.0.0.0/24 1.0.0.0/24]
}

func ExampleIPv4NetList_Sort() {
	nets := []string{"10.0.0.0/24", "1.0.0.0/24", "10.0.0.0/8", "192.168.1.0/26", "8.8.8.8/32", "10.0.0.0/8"}
	list, _ := NewIPv4NetList(nets)
	list.Sort()
	fmt.Println(list)
	// Output: [1.0.0.0/24 8.8.8.8/32 10.0.0.0/24 10.0.0.0/8 10.0.0.0/8 192.168.1.0/26]
}

func ExampleIPv4NetList_Summ() {
	nets := []string{"10.0.0.0/24", "10.0.0.64/26", "1.1.1.0/24", "1.0.0.0/8", "3.4.5.6/32", "3.4.5.8/31", "2.2.2.224/27"}
	list, _ := NewIPv4NetList(nets)
	list = list.Summ()
	fmt.Println(list)
	// Output: [1.0.0.0/8 2.2.2.224/27 3.4.5.6/32 3.4.5.8/31 10.0.0.0/24]
}

func Test_NewIPv4NetList(t *testing.T) {
	cases := []struct {
		given []string
		err   bool
	}{
		{
			[]string{"10.0.0.0/24", "1.0.0.0/24"},
			false,
		},
		{
			[]string{"10.0.0.0/24", "1.0.0.0/24", "1/x"},
			true,
		},
	}

	for _, c := range cases {
		_, err := NewIPv4NetList(c.given)
		if err != nil {
			if !c.err {
				t.Errorf("NewIPv4NetList(%s) unexpected error: %s", c.given, err.Error())
			}
		} else if c.err {
			t.Errorf("NewIPv4NetList(%s) expected error but none raised", c.given)
		}
	}
}

func Test_IPv4NetList_Summ(t *testing.T) {
	cases := []struct {
		given  []string
		expect []string
	}{
		{ // summarize a complete subnet
			[]string{"10.0.0.0/29", "10.0.0.8/30", "10.0.0.12/30", "10.0.0.16/28", "10.0.0.32/27", "10.0.0.64/26", "10.0.0.128/25"},
			[]string{"10.0.0.0/24"},
		},
		{ // summarize the complete ip space
			[]string{"10.0.0.0/24", "1.0.0.0/8", "3.4.5.6/32", "3.4.5.8/31", "0.0.0.0/0"},
			[]string{"0.0.0.0/0"},
		},
		{ // mix of out of order, duplicates, non-contiguous, subnet-of
			[]string{"10.0.1.0/25", "10.0.1.0/26", "10.0.0.16/28", "10.0.0.32/27", "10.0.0.128/26", "10.0.0.192/26", "10.0.0.32/27"},
			[]string{"10.0.0.16/28", "10.0.0.32/27", "10.0.0.128/25", "10.0.1.0/25"},
		},
	}

	for _, c := range cases {
		list, err := NewIPv4NetList(c.given)
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
