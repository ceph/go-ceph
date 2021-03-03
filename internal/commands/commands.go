package commands

import (
	"encoding/json"

	"github.com/ceph/go-ceph/rados"
)

// MgrCommander in an interface for the API needed to execute JSON formatted
// commands on the ceph mgr.
type MgrCommander interface {
	MgrCommand(buf [][]byte) ([]byte, string, error)
}

// MonCommander is an interface for the API needed to execute JSON formatted
// commands on the ceph mon(s).
type MonCommander interface {
	MonCommand(buf []byte) ([]byte, string, error)
}

// RadosCommander provides an interface for APIs needed to execute JSON
// formatted commands on the Ceph cluster.
type RadosCommander interface {
	MgrCommander
	MonCommander
}

func validate(m interface{}) error {
	if m == nil {
		return rados.ErrNotConnected
	}
	return nil
}

// RawMgrCommand takes a byte buffer and sends it to the MGR as a command.
// The buffer is expected to contain preformatted JSON.
func RawMgrCommand(m MgrCommander, buf []byte) Response {
	if err := validate(m); err != nil {
		return Response{err: err}
	}
	return NewResponse(m.MgrCommand([][]byte{buf}))
}

// MarshalMgrCommand takes an generic interface{} value, converts it to JSON
// and sends the json to the MGR as a command.
func MarshalMgrCommand(m MgrCommander, v interface{}) Response {
	b, err := json.Marshal(v)
	if err != nil {
		return Response{err: err}
	}
	return RawMgrCommand(m, b)
}

// RawMonCommand takes a byte buffer and sends it to the MON as a command.
// The buffer is expected to contain preformatted JSON.
func RawMonCommand(m MonCommander, buf []byte) Response {
	if err := validate(m); err != nil {
		return Response{err: err}
	}
	return NewResponse(m.MonCommand(buf))
}

// MarshalMonCommand takes an generic interface{} value, converts it to JSON
// and sends the json to the MGR as a command.
func MarshalMonCommand(m MonCommander, v interface{}) Response {
	b, err := json.Marshal(v)
	if err != nil {
		return Response{err: err}
	}
	return RawMonCommand(m, b)
}
