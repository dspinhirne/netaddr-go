# netaddr
A Go library for performing calculations on IPv4 and IPv6 subnets. There is also limited support for EUI addresses.


# Installation
It should be noted that while the repository name is "netaddr-go" the package name is simply "netaddr".
	go get github.com/dspinhirne/netaddr-go


# Usage
	import "github.com/dspinhirne/netaddr-go"

	net,_ := netaddr.NewIPv4Net("192.168.1.0/24")


# Documentation
Available online [here](https://godoc.org/github.com/dspinhirne/netaddr-go).


# Current State
Finalizing for an official 1.0 release. Design is not 100% guaranteed to be free of changes until the "1.0" branch is created.
