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
	CephUnknown
)

// CurrentCephVersion is the current Ceph version
func CurrentCephVersion() CephVersion {
	vname := os.Getenv("CEPH_VERSION")
	return CephVersionOfString(vname)
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
	default:
		return CephUnknown
	}
}
