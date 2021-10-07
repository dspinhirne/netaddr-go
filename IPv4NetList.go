package netaddr

import (
	"fmt"
	"sort"
)

// IPv4NetList is a slice of IPv4 types
type IPv4NetList []*IPv4Net

// NewIPv4NetList parses a slice of IP networks into a IPv4NetList.
func NewIPv4NetList(networks []string) (IPv4NetList, error) {
	list := make(IPv4NetList, len(networks), len(networks))
	for i, e := range networks {
		net, err := ParseIPv4Net(e)
		if err != nil {
			return nil, fmt.Errorf("Error parsing item index %d. %s", i, err.Error())
		}
		list[i] = net
	}
	return list, nil
}

// Len is used to implement the sort interface
func (list IPv4NetList) Len() int { return len(list) }

// Less is used to implement the sort interface
func (list IPv4NetList) Less(i, j int) bool {
	cmp, _ := list[i].Cmp(list[j])
	return cmp == -1
}

// Sort sorts the list using sort.Sort(). Returns itself.
func (list IPv4NetList) Sort() IPv4NetList {
	sort.Sort(list)
	return list
}

// Summ returns a copy of the list with the contained IPv4Net entries
// sorted and summarized as much as possible.
func (list IPv4NetList) Summ() IPv4NetList {
	var summd IPv4NetList
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
func (list IPv4NetList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }

// NON EXPORTED

// discardSubnets returns a sorted copy of the IPv4NetList with
// any entries which are subnets of other entries removed.
func (list IPv4NetList) discardSubnets() IPv4NetList {
	keepers := IPv4NetList{}
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
			keepers = append(IPv4NetList{last}, keepers...)
		}
	}
	return keepers
}

// summPeers returns a copy of the IPv4NetList with any
// merge-able subnets Summ'd together.
func (list IPv4NetList) summPeers() IPv4NetList {
	summd := list.Sort()
	for {
		listLen := len(summd)
		last := listLen - 1
		var tmpList IPv4NetList
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
