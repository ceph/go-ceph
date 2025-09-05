//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
)

// bindAddr is the low level type in the smb mgr module json/yaml and acts
// a bit like a union where the key determines the type. To avoid callers
// setting more than one field at once this is a private type mediated
// thru BindAddress.
type bindAddr struct {
	Address string `json:"address,omitempty"`
	Network string `json:"network,omitempty"`
}

// BindAddress indicates an IP Address or Network (IP address range) that
// the cluster is permitted to listen to.
type BindAddress struct {
	a bindAddr
}

// NewBindAddress returns a BindAddress associated with a single IP address.
func NewBindAddress(s string) BindAddress {
	var b BindAddress
	b.a.Address = s
	return b
}

// NewNetworkBindAddress returns a BindAddress associated with a network
// address. The network address is a range of addresses determined using
// a prefix length.
func NewNetworkBindAddress(s string) BindAddress {
	var b BindAddress
	b.a.Network = s
	return b
}

// MarshalJSON supports marshalling a BindAddress to JSON.
func (b BindAddress) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.a)
}

// UnmarshalJSON supports un-marshalling a JSON to a BindAddress.
func (b *BindAddress) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &b.a)
}

// Address returns the IP address value.
func (b BindAddress) Address() string {
	return b.a.Address
}

// Network returns the IP network address value.
func (b BindAddress) Network() string {
	return b.a.Network
}

// IsNetwork returns true if this BindAddress contains a network value.
func (b BindAddress) IsNetwork() bool {
	return b.a.Network != ""
}

// String returns a string representing this BindAddress.
func (b BindAddress) String() string {
	if b.a.Network != "" {
		return "network:" + b.a.Network
	}
	return "address:" + b.a.Address
}
