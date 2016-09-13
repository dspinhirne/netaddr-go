package netaddr

import "testing"
import "fmt"

func ExampleIPv4List_Sort() {
	ips := []string{"10.0.0.0", "1.0.0.0", "10.0.0.0", "192.168.1.0", "8.8.8.8", "10.0.0.0"}
	list, _ := NewIPv4List(ips)
	list.Sort()
	fmt.Println(list)
	// Output: [1.0.0.0 8.8.8.8 10.0.0.0 10.0.0.0 10.0.0.0 192.168.1.0]
}

func Test_NewIPv4List(t *testing.T) {
	cases := []struct {
		given []string
		err   bool
	}{
		{
			[]string{"10.0.0.0", "1.0.0.0"},
			false,
		},
		{
			[]string{"10.0.0.0", "1.0.0.0", "1"},
			true,
		},
	}

	for _, c := range cases {
		_, err := NewIPv4List(c.given)
		if err != nil {
			if !c.err {
				t.Errorf("NewIPv4List(%s) unexpected error: %s", c.given, err.Error())
			}
		} else if c.err {
			t.Errorf("NewIPv4List(%s) expected error but none raised", c.given)
		}
	}
}
