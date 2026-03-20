//go:build !(pacific || quincy) && ceph_preview

package builder

import (
	ccom "github.com/ceph/go-ceph/common/commands"

	"github.com/ceph/go-ceph/internal/commands"
)

var queryCommand = []byte(`{"prefix": "get_command_descriptions"}`)

func queryMgr(m ccom.MgrCommander) commands.Response {
	return commands.NewResponse(m.MgrCommand([][]byte{queryCommand}))
}

func queryMon(m ccom.MonCommander) commands.Response {
	return commands.NewResponse(m.MonCommand(queryCommand))
}

// QueryMgrJSON makes a request to the Ceph MGR to describe the commands that
// the service knows about. This function returns the response as raw JSON
// encoded bytes.
func QueryMgrJSON(m ccom.MgrCommander) ([]byte, error) {
	response := queryMgr(m).NoStatus()
	if response.Ok() {
		return response.Body(), nil
	}
	return nil, response
}

// QueryMonJSON makes a request to the Ceph MON to describe the commands that
// the service knows about. This function returns the response as raw JSON
// encoded bytes.
func QueryMonJSON(m ccom.MonCommander) ([]byte, error) {
	response := queryMon(m).NoStatus()
	if response.Ok() {
		return response.Body(), nil
	}
	return nil, response
}

// QueryMgrDescriptions makes a request to the Ceph MGR to describe the
// commands that the service knows about. This function returns the response as
// a CommandDescriptions object.
func QueryMgrDescriptions(m ccom.MgrCommander) (CommandDescriptions, error) {
	cd := CommandDescriptions{}
	if err := queryMgr(m).NoStatus().Unmarshal(&cd).End(); err != nil {
		return CommandDescriptions{}, err
	}
	return cd, nil
}

// QueryMonDescriptions makes a request to the Ceph MON to describe the
// commands that the service knows about. This function returns the response as
// a CommandDescriptions object.
func QueryMonDescriptions(m ccom.MonCommander) (CommandDescriptions, error) {
	cd := CommandDescriptions{}
	if err := queryMon(m).NoStatus().Unmarshal(&cd).End(); err != nil {
		return CommandDescriptions{}, err
	}
	return cd, nil
}
