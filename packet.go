package vrrp

type Packet struct {
	VersionAndType     uint8
	ID                 uint8
	Priority           uint8
	NAddr              uint8
	RsvdAndMaxAdverInt uint16
	Checksum           uint16
}
