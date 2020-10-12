package vrrp

import (
	"errors"
	"net"
)

var (
	ErrExist      = errors.New("virtual router exists already")
	ErrNotOwnAddr = errors.New("interface does not own the address")
	ErrAddrInuse  = errors.New("the address(s) used")
)

func NewVirtualRouter(VRID byte, priority byte, flags byte, ifname string, addrs ...string) (vr *VirtualRouter, err error) {
	var ifc *net.Interface
	if ifc, err = net.InterfaceByName(ifname); err != nil {
		return
	}

	vr = newVirtualRouter(VRID, priority, flags, ifc)
	for _, addr := range addrs {
		vr.AddAddr(addr)
	}
	return
}
