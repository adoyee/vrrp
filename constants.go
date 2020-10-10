package vrrp

import (
	"net"
	"time"
)

const (
	IPv4 = 0x1
	IPv6 = 0x2
)

type StateType byte

const (
	stateInit StateType = iota
	stateMaster
	stateBackup
	stateDown
)

func (state StateType) String() string {
	switch state {
	case stateInit:
		return "Initialize"
	case stateMaster:
		return "Master"
	case stateBackup:
		return "Backup"
	case stateDown:
		return "Interface Down"
	default:
		return "Unknown state"
	}
}

const (
	vrrpMultiTTL       = 255
	vrrpProtocolNumber = 112
	maxPriority        = 255
)

var (
	multiAddrIPv4 = net.IP{224, 0, 0, 18}
	multiAddrIPv6 = net.ParseIP("FF02:0:0:0:0:0:0:12")
)

var (
	defaultPreempt               = true
	defaultPriority              = 100
	defaultAdvertisementInterval = 1 * time.Second
)

const (
	FlagPreempt = 0x1
	FlagAccept  = 0x2
	FlagUnicast = 0x4
	FlagIPv6    = 0x8
	FlagDefault = FlagPreempt | FlagAccept | FlagUnicast
)
