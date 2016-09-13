package netaddr

import "testing"
import "fmt"

func ExampleIPv6List_Sort() {
	ips := []string{"1::", "2::", "1::1", "::", "::1", "fec0::1"}
	list, _ := NewIPv6List(ips)
	list.Sort()
	fmt.Println(list)
	// Output: [:: ::1 1:: 1::1 2:: fec0::1]
}

func Test_NewIPv6List(t *testing.T) {
	cases := []struct {
		given []string
		err   bool
	}{
		{
			[]string{"1::", "2::", "::"},
			false,
		},
		{
			[]string{"1::", "2::", "::", "fec0"},
			true,
		},
	}

	for _, c := range cases {
		_, err := NewIPv6List(c.given)
		if err != nil {
			if !c.err {
				t.Errorf("NewIPv6List(%s) unexpected error: %s", c.given, err.Error())
			}
		} else if c.err {
			t.Errorf("NewIPv6List(%s) expected error but none raised", c.given)
		}
	}
}
