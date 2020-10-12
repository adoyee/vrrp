package vrrp

import (
	"net"
	"testing"
)

func TestInterfaceOwnIP(t *testing.T) {
	ifc, _ := net.InterfaceByName("lo")
	ip := net.IP{127, 0, 0, 1}
	own := interfaceOwnIP(ifc, ip)
	if !own {
		t.Fatal("loopback interface should own 127.0.0.1")
	}
}
