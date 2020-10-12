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

type vrManager struct {
	interfaces []*net.Interface
	vrIPv6     []*VirtualRouter
	vrIPv4     []*VirtualRouter
	vrs        map[uint32]*VirtualRouter
	allAddr    []net.IP
	lock       sync.Mutex
}

var mgr *vrManager

func init() {
	mgr = &vrManager{}
	mgr.interfaces = make([]*net.Interface, 0, 8)
	mgr.allAddr = make([]net.IP, 0, 8)
	mgr.vrIPv4 = make([]*VirtualRouter, 0, 8)
	mgr.vrIPv6 = make([]*VirtualRouter, 0, 8)
	mgr.vrs = make(map[uint32]*VirtualRouter)
}

func (mgr *vrManager) addVirtualRouter(vr *VirtualRouter) (err error) {
	mgr.lock.Lock()
	defer mgr.lock.Unlock()

	if mgr.vrIsExist(vr.vrid, vr.ifc, vr.flags) {
		return ErrExist
	}

	if vr.Priority() == maxPriority {
		if !mgr.validAddrOwner(vr) {
			return ErrNotOwnAddr
		}
	}

	for _, ip := range vr.vrAddrs {
		if mgr.validAddrInuse(ip) {
			return ErrAddrInuse
		}
	}

	if err = mgr.enableInterfaceMulticast(vr.ifc); err != nil {
		return
	}

	if vr.ipv6 {
		mgr.vrIPv6 = append(mgr.vrIPv6, vr)
	} else {
		mgr.vrIPv4 = append(mgr.vrIPv4, vr)
	}

	var mapid uint32
	mapid = uint32(vr.ifc.Index) << 16
	mapid |= uint32(vr.Priority())
	mgr.vrs[mapid] = vr

	return
}

func (mgr *vrManager) vrIsExist(vrid byte, ifc *net.Interface, ipvx byte) bool {
	var vrs []*VirtualRouter
	if ipvx&FlagIPv6 != 0 {
		vrs = mgr.vrIPv6
	} else {
		vrs = mgr.vrIPv4
	}

	for _, vr := range vrs {
		if vr.ifc.Index == ifc.Index && vr.VRID() == vrid {
			return true
		}
	}

	return false
}

func (mgr *vrManager) validAddrOwner(vr *VirtualRouter) bool {
	ips := vr.vrAddrs
	for _, ip := range ips {
		if !interfaceOwnIP(vr.ifc, ip) {
			return false
		}
	}
	return true
}

func (mgr *vrManager) validAddrInuse(ip net.IP) bool {
	for _, iip := range mgr.allAddr {
		if iip.String() == ip.String() {
			return true
		}
	}
	return false
}

func (mgr *vrManager) enableInterfaceMulticast(ifc *net.Interface) (err error) {
	for _, iifc := range mgr.interfaces {
		if ifc.Index == iifc.Index {
			return nil
		}
	}
	err = enableInterfaceMulticast(ifc)
	if err != nil {
		return
	}

	mgr.interfaces = append(mgr.interfaces, ifc)
	return
}

func interfaceOwnIP(ifc *net.Interface, ip net.IP) bool {
	addrs, _ := ifc.Addrs()
	for _, addr := range addrs {
		iip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return false
		}
		if ip.String() == iip.String() {
			return true
		}
	}
	return false
}
