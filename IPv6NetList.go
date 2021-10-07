package netaddr

import (
	"fmt"
	"sort"
)

// IPv6NetList is a slice of IPv6 types
type IPv6NetList []*IPv6Net

// NewIPv6NetList parses a slice of IP networks into a IPv6NetList.
func NewIPv6NetList(networks []string) (IPv6NetList, error) {
	list := make(IPv6NetList, len(networks), len(networks))
	for i, e := range networks {
		net, err := ParseIPv6Net(e)
		if err != nil {
			return nil, fmt.Errorf("Error parsing item index %d. %s", i, err.Error())
		}
		list[i] = net
	}
	return list, nil
}

// Len is used to implement the sort interface
func (list IPv6NetList) Len() int { return len(list) }

// Less is used to implement the sort interface
func (list IPv6NetList) Less(i, j int) bool {
	cmp, _ := list[i].Cmp(list[j])
	return cmp == -1
}

// Sort sorts the list using sort.Sort(). Returns itself.
func (list IPv6NetList) Sort() IPv6NetList {
	sort.Sort(list)
	return list
}

// Summ returns a copy of the list with the contained IPv6Net entries
// sorted and summarized as much as possible.
func (list IPv6NetList) Summ() IPv6NetList {
	var summd IPv6NetList
	if len(list) > 1 {
		summd = list.discardSubnets()
	} else if len(list) == 1 {
		summd = append(summd, list...)
	}

	if len(summd) > 1 {
		summd = summd.summPeers()
	}
	return summd
}

// Swap is used to implement the sort interface
func (list IPv6NetList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

// NON EXPORTED

// discardSubnets returns a sorted copy of the IPv6NetList with
// any entries which are subnets of other entries removed.
func (list IPv6NetList) discardSubnets() IPv6NetList {
	keepers := IPv6NetList{}
	if len(list)>0{ // only do work if we have something to work on
		last := list[len(list)-1]
		keepLast := true
		for _, e := range list {
			isRel, rel := last.Rel(e)
			if !isRel { // keep unrelated nets
				keepers = append(keepers, e)
			} else if isRel && rel == -1 { // keep supernets but do not keepLast
				keepers = append(keepers, e)
				keepLast = false
			}
		}

		if len(keepers) > 0 {
			keepers = keepers.discardSubnets()
		}
		if keepLast{
			keepers = append(IPv6NetList{last}, keepers...)
		}
	}
	return keepers
}

// summPeers returns a copy of the IPv6NetList with any
// merge-able subnets Summ'd together.
func (list IPv6NetList) summPeers() IPv6NetList {
	summd := list.Sort()
	for {
		listLen := len(summd)
		last := listLen - 1
		var tmpList IPv6NetList
		for i := 0; i < listLen; i += 1 {
			net := summd[i]
			next := i + 1
			if i != last {
				// if we can summarize 2 consecutive entries then store the new
				// summary net and discard the 2 original networks
				newNet := net.Summ(summd[next])
				if newNet != nil { // can summarize. keep summary net
					tmpList = append(tmpList, newNet)
					i += 1 // skip over the next entry
				} else { // cant summarize. keep existing
					tmpList = append(tmpList, net)
				}
			} else {
				tmpList = append(tmpList, net) // keep last
			}
		}
		// stop if summd is not getting shorter
		if len(tmpList) == listLen {
			break
		}
		summd = tmpList
	}
	return summd
}
