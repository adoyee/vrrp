package vrrp

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type VirtualRouter struct {
	vrid     byte
	priority byte
	vrAddrs  []net.IP

	advertisementInterval         uint16
	advertisementIntervalOfMaster uint16
	masterDownInterval            uint16

	state          StateType
	flags          byte
	preempt        bool
	accept         bool
	ipv6           bool
	unicast        bool
	ifc            *net.Interface
	skewTime       uint16
	lock           sync.Mutex
	macAddressIPv4 net.HardwareAddr
	macAddressIPv6 net.HardwareAddr
}

func NewVirtualRouterOwnInterface(VRID byte, ifname string, IPvX byte) (vr *VirtualRouter, err error) {
	var ifc *net.Interface
	if ifc, err = net.InterfaceByName(ifname); err != nil {
		return
	}

	vr = newVirtualRouter(VRID, maxPriority, FlagDefault, ifc)
	return
}

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

func newVirtualRouter(VRID byte, priority byte, flags byte, ifc *net.Interface) (vr *VirtualRouter) {
	vr = &VirtualRouter{
		vrid:     VRID,
		priority: priority,
		ifc:      ifc,
	}

	vr.macAddressIPv4, _ = net.ParseMAC(fmt.Sprintf("00-00-5E-00-01-%X", VRID))
	vr.macAddressIPv6, _ = net.ParseMAC(fmt.Sprintf("00-00-5E-00-02-%X", VRID))
	vr.state = stateInit
	vr.SetAdvInterval(defaultAdvertisementInterval)
	vr.flags = flags
	vr.preempt = flags&FlagPreempt != 0
	vr.accept = flags&FlagAccept != 0
	vr.ipv6 = flags&FlagIPv6 != 0
	vr.unicast = flags&FlagUnicast != 0
	vr.vrAddrs = make([]net.IP, 0, 32)

	return
}

func (vr *VirtualRouter) State() StateType {
	return vr.state
}

func (vr *VirtualRouter) Priority() byte {
	return vr.priority
}

func (vr *VirtualRouter) VRID() byte {
	return vr.vrid
}

func (vr *VirtualRouter) AcceptModeEnabled() bool {
	return vr.accept
}

func (vr *VirtualRouter) IsUnicast() bool {
	return vr.unicast
}

func (vr *VirtualRouter) IsOwner() bool {
	return vr.priority == maxPriority
}

func (vr *VirtualRouter) NAddr() int {
	vr.lock.Lock()
	defer vr.lock.Unlock()
	return len(vr.vrAddrs)
}

func (vr *VirtualRouter) SetAdvInterval(interval time.Duration) {
	if interval < time.Millisecond*100 {
		interval = time.Millisecond * 100
	}
	vr.advertisementInterval = uint16(interval / (time.Millisecond * 10))
}

func (vr *VirtualRouter) AddAddr(addr string) {
	ip := net.ParseIP(addr)
	if ip != nil {
		vr.AddIPvXAddr(ip)
	}
}

func (vr *VirtualRouter) RemoveAddr(addr string) {
	ip := net.ParseIP(addr)
	if ip != nil {
		vr.RemoveIPvxAddr(ip)
	}
}

func (vr *VirtualRouter) AddIPvXAddr(addr net.IP) {

}

func (vr *VirtualRouter) RemoveIPvxAddr(addr net.IP) {

}

func (vr *VirtualRouter) Start() (err error) {
	return
}

func (vr *VirtualRouter) Stop() {
	return
}

func (vr *VirtualRouter) addInterfaceAddr() {
	addresses, err := vr.ifc.Addrs()
	if err != nil {
		return
	}

	for _, addr := range addresses {
		a := addr.String()
		slash := strings.Index(a, "/")
		if slash != -1 {
			a = a[:slash]
		}
		ip := net.ParseIP(a)
		if ip == nil {
			continue
		}

		if !ip.IsGlobalUnicast() {
			continue
		}

		vr.AddIPvXAddr(ip)
	}
}
