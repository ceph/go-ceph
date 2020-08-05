package admin

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ceph/go-ceph/rados"
)

// RadosCommander provides an interface to execute JSON-formatted commands that
// allow the cephfs administrative functions to interact with the Ceph cluster.
type RadosCommander interface {
	MgrCommand(buf [][]byte) ([]byte, string, error)
}

// FSAdmin is used to administrate CephFS within a ceph cluster.
type FSAdmin struct {
	conn RadosCommander
}

// New creates an FSAdmin automatically based on the default ceph
// configuration file. If more customization is needed, create a
// *rados.Conn as you see fit and use NewFromConn to use that
// connection with these administrative functions.
func New() (*FSAdmin, error) {
	conn, err := rados.NewConn()
	if err != nil {
		return nil, err
	}
	err = conn.ReadDefaultConfigFile()
	if err != nil {
		return nil, err
	}
	err = conn.Connect()
	if err != nil {
		return nil, err
	}
	return NewFromConn(conn), nil
}

// NewFromConn creates an FSAdmin management object from a preexisting
// rados connection. The existing connection can be rados.Conn or any
// type implementing the RadosCommander interface. This may be useful
// if the calling layer needs to inject additional logging, error handling,
// fault injection, etc.
func NewFromConn(conn RadosCommander) *FSAdmin {
	return &FSAdmin{conn}
}

func (fsa *FSAdmin) validate() error {
	if fsa.conn == nil {
		return rados.ErrNotConnected
	}
	return nil
}

// rawMgrCommand takes a byte buffer and sends it to the MGR as a command.
// The buffer is expected to contain preformatted JSON.
func (fsa *FSAdmin) rawMgrCommand(buf []byte) ([]byte, string, error) {
	if err := fsa.validate(); err != nil {
		return nil, "", err
	}
	return fsa.conn.MgrCommand([][]byte{buf})
}

// marshalMgrCommand takes an generic interface{} value, converts it to JSON and
// sends the json to the MGR as a command.
func (fsa *FSAdmin) marshalMgrCommand(v interface{}) ([]byte, string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, "", err
	}
	return fsa.rawMgrCommand(b)
}

type listNamedResult struct {
	Name string `json:"name"`
}

func parseListNames(res []byte, status string, err error) ([]string, error) {
	var r []listNamedResult
	if err := unmarshalResponseJSON(res, status, err, &r); err != nil {
		return nil, err
	}
	vl := make([]string, len(r))
	for i := range r {
		vl[i] = r[i].Name
	}
	return vl, nil
}

// checkEmptyResponseExpected returns an error if the result or status
// are non-empty.
func checkEmptyResponseExpected(res []byte, status string, err error) error {
	if err != nil {
		return err
	}
	if len(res) != 0 {
		return fmt.Errorf("unexpected response: %s", string(res))
	}
	if status != "" {
		return fmt.Errorf("error status: %s", status)
	}
	return nil
}

func unmarshalResponseJSON(res []byte, status string, err error, v interface{}) error {
	if err != nil {
		return err
	}
	if status != "" {
		return fmt.Errorf("error status: %s", status)
	}
	return json.Unmarshal(res, v)
}

// modeString converts a unix-style mode value to a string-ified version in an
// octal representation (e.g. "777", "700", etc). This format is expected by
// some of the ceph JSON command inputs.
func modeString(m int, force bool) string {
	if force || m != 0 {
		return strconv.FormatInt(int64(m), 8)
	}
	return ""
}

// uint64String converts a uint64 to a string. Some of the ceph json commands
// can take a string or "int" (as a string). This is a common function for
// doing that conversion.
func uint64String(v uint64) string {
	return strconv.FormatUint(uint64(v), 10)
}
