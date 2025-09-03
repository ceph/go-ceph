//go:build !octopus && ceph_preview

package osd

import (
	"bytes"
	"fmt"
	"net"
	"strings"

	"github.com/ceph/go-ceph/internal/commands"
)

// OSDBlocklist returns the list of blocklisted clients.
//
// Similar To:
//
//	ceph osd blocklist ls
func (osda *Admin) OSDBlocklist() ([]string, error) {
	resp := []string{}

	cmd := map[string]string{"prefix": "osd blocklist ls"}

	buf := commands.MarshalMonCommand(osda.conn, cmd)
	if !buf.Ok() {
		return resp, buf.End()
	}

	resp = strings.FieldsFunc(string(buf.Body()), func(r rune) bool {
		return r == '\n'
	})

	return resp, nil
}

// AddressEntry contains the ip or network address string along with the
// optional expire value in seconds.
type AddressEntry struct {
	addr   string
	expire float64
}

func isValidIP(addr string) bool {
	if ip := net.ParseIP(addr); ip != nil {
		return true
	}

	return false
}

func isValidCIDR(addr string) bool {
	if _, _, err := net.ParseCIDR(addr); err == nil {
		return true
	}

	return false
}

// OSDBlocklistAdd adds an ip address or network address in CIDR format to the
// blocklist.
//
// Similar To:
//
//	ceph osd blocklist [range] add <ip_addr|cidr_network> [expire]
func (osda *Admin) OSDBlocklistAdd(entry AddressEntry) (string, error) {
	if entry.addr == "" {
		return "", ErrEmptyArgument
	}

	cmd := []byte(`{"prefix": "osd blocklist","blocklistop": "add",`)

	if entry.expire < 0 {
		return "", ErrInvalidArgument
	}

	if entry.expire != 0 {
		cmd = append(cmd, []byte(`"expire": entry.expire,`)...)
		cmd = bytes.Replace(cmd,
			[]byte("entry.expire"),
			[]byte(fmt.Sprintf("%.1f", entry.expire)),
			-1,
		)
	}

	if !isValidIP(entry.addr) {
		if !isValidCIDR(entry.addr) {
			return "", ErrInvalidArgument
		}
		cmd = append(cmd, []byte(`"range": "range",`)...)
	}

	cmd = append(cmd, []byte(`"addr": "entry.addr",}`)...)
	cmd = bytes.Replace(cmd,
		[]byte("entry.addr"),
		[]byte(entry.addr),
		-1,
	)

	buf := commands.RawMonCommand(osda.conn, cmd)

	return buf.Status(), buf.End()
}

// OSDBlocklistRemove removes an ip address or network address from the
// blocklist.
//
// Similar To:
//
//	ceph osd blocklist [range] rm <ip_addr|cidr_network>
func (osda *Admin) OSDBlocklistRemove(entry AddressEntry) (string, error) {
	if entry.addr == "" {
		return "", ErrEmptyArgument
	}

	cmd := map[string]string{"prefix": "osd blocklist",
		"blocklistop": "rm",
		"addr":        entry.addr,
	}

	if !isValidIP(entry.addr) {
		if !isValidCIDR(entry.addr) {
			return "", ErrInvalidArgument
		}
		cmd["range"] = "range"
	}

	buf := commands.MarshalMonCommand(osda.conn, cmd)
	if buf.Ok() && !strings.Contains(buf.Status(), "un-blocklisting") {
		return buf.Status(), ErrNotFound
	}

	return buf.Status(), buf.End()
}
