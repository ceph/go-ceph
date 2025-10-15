//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import "fmt"

// RemoteControl configures the optional smb cluster remote control subsystem.
type RemoteControl struct {
	// Enabled is used to explicitly enable or disable the remote control
	// subsystem. If Enabled is true the remote control subsystem will be
	// enabled, if false always disabled. If unset (nil) the state of the
	// subsystem is determined by the TLS sources being set or not.
	Enabled *bool `json:"enabled,omitempty"`
	// Cert is used to provide a TLS certificate to the remote control service.
	Cert *TLSCredentialSource `json:"cert,omitempty"`
	// Key is used to provide a TLS key to the remote control service.
	Key *TLSCredentialSource `json:"key,omitempty"`
	// CACert is used to provide a TLS CA certificate to the remote control
	// service.
	CACert *TLSCredentialSource `json:"ca_cert,omitempty"`
}

// Validate returns an error describing an issue with the remote control
// configuration or nil if the resource object is valid.
func (rc *RemoteControl) Validate() error {
	hasCert := rc.Cert != nil
	hasKey := rc.Key != nil
	hasCACert := rc.CACert != nil
	if hasCert != hasKey {
		return fmt.Errorf("cert and key must be specified together")
	}
	if hasCACert != hasCert {
		return fmt.Errorf("a CA cert must be specified with a cert and key")
	}
	return nil
}
