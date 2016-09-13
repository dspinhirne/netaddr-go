package netaddr

import (
	"fmt"
	"sort"
)

// NewIPv6List parses a slice of IPv6 addresses into a IPv6List.
func NewIPv6List(ips []string) (IPv6List, error) {
	list := make(IPv6List, len(ips), len(ips))
	for i, e := range ips {
		ip, err := ParseIPv6(e)
		if err != nil {
			return nil, fmt.Errorf("Error parsing item index %d. %s", i, err.Error())
		}
		list[i] = ip
	}
	return list, nil
}

// IPv6List is a slice of IPv6 types
type IPv6List []*IPv6

// Len is used to implement the sort interface
func (list IPv6List) Len() int { return len(list) }

// Less is used to implement the sort interface
func (list IPv6List) Less(i, j int) bool {
	cmp, _ := list[i].Cmp(list[j])
	return cmp == -1
}

// Sort sorts the list using sort.Sort(). Returns itself.
func (list IPv6List) Sort() IPv6List {
	sort.Sort(list)
	return list
}

// Swap is used to implement the sort interface
func (list IPv6List) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
