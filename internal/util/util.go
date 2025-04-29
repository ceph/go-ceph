package util

import "os"

// CephVersion type
type CephVersion int

// Enum of known CephVersions
const (
	CephNautilus CephVersion = 14 + iota
	CephOctopus
	CephPacific
	CephQuincy
	CephReef
	CephSquid
	CephTentacle
	CephUnknown
)

// List of known CephVersion strings
const (
	Nautilus = "nautilus"
	Octopus  = "octopus"
	Pacific  = "pacific"
	Quincy   = "quincy"
	Reef     = "reef"
	Squid    = "squid"
	Tentacle = "tentacle"
	Main     = "main"
)

// CurrentCephVersion is the current Ceph version
func CurrentCephVersion() CephVersion {
	vname := os.Getenv("CEPH_VERSION")
	return CephVersionOfString(vname)
}

// CurrentCephVersionString is the current Ceph version string
func CurrentCephVersionString() string {
	switch vname := os.Getenv("CEPH_VERSION"); vname {
	case Nautilus, Octopus, Pacific, Quincy, Reef, Squid, Tentacle, Main:
		return vname
	}
	return ""
}

// CephVersionOfString converts a string to CephVersion
func CephVersionOfString(vname string) CephVersion {
	switch vname {
	case "nautilus":
		return CephNautilus
	case "octopus":
		return CephOctopus
	case "pacific":
		return CephPacific
	case "quincy":
		return CephQuincy
	case "reef":
		return CephReef
	case "squid":
		return CephSquid
	case "tentacle":
		return CephTentacle
	default:
		return CephUnknown
	}
}
