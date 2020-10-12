package vrrp

import (
	"net"
	"sync"
)

type vrrpInterface struct {
	ifc     *net.Interface
	enabled bool

	lock sync.Mutex
}

func (ifc *vrrpInterface) Index() int {
	return ifc.ifc.Index
}

func (ifc *vrrpInterface) Name() string {
	return ifc.ifc.Name
}

func (ifc *vrrpInterface) HardWareAddr() net.HardwareAddr {
	return ifc.ifc.HardwareAddr
}

func (ifc *vrrpInterface) enableMulticast() (err error) {
	ifc.lock.Lock()
	defer ifc.lock.Unlock()
	if ifc.enabled {
		return
	}

	return
}

func (ifc *vrrpInterface) disableMulticast() (err error) {
	ifc.lock.Lock()
	defer ifc.lock.Unlock()
	if !ifc.enabled {
		return
	}
	return
}

type globalVariables struct {
	interfaces []*vrrpInterface
	vrIPv6     [256]*VirtualRouter
	vrIPv4     [256]*VirtualRouter
	lock       sync.Mutex
}
