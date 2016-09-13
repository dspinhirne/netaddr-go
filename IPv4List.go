package netaddr

import (
	"fmt"
	"sort"
)

// NewIPv4List parses a slice of IPv4 addresses into a IPv4List.
func NewIPv4List(ips []string) (IPv4List, error) {
	list := make(IPv4List, len(ips), len(ips))
	for i, e := range ips {
		ip, err := ParseIPv4(e)
		if err != nil {
			return nil, fmt.Errorf("Error parsing item index %d. %s", i, err.Error())
		}
		list[i] = ip
	}
	return list, nil
}

// IPv4List is a slice of IPv4 types
type IPv4List []*IPv4

// Len is used to implement the sort interface
func (list IPv4List) Len() int { return len(list) }

// Less is used to implement the sort interface
func (list IPv4List) Less(i, j int) bool {
	cmp, _ := list[i].Cmp(list[j])
	return cmp == -1
}

// Sort sorts the list using sort.Sort(). Returns itself.
func (list IPv4List) Sort() IPv4List {
	sort.Sort(list)
	return list
}

// Swap is used to implement the sort interface
func (list IPv4List) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
