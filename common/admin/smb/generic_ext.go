//go:build !(pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"

	"github.com/ceph/go-ceph/internal/commands"
)

type genericResultCommon struct {
	Resource GenericResource `json:"resource"`
	Message  string          `json:"msg"`
	Success  bool            `json:"success"`
	State    string          `json:"state"`
}

type genericResult struct {
	genericResultCommon
	status map[string]any
}

func (gr *genericResult) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &gr.genericResultCommon); err != nil {
		return err
	}
	gr.status = map[string]any{}
	return json.Unmarshal(data, &gr.status)
}

type genericResultGroup struct {
	Success bool            `json:"success"`
	Results []genericResult `json:"results"`
}

func applyGenericUnmarshal(r commands.Response) (ResultGroup, error) {
	rg := genericResultGroup{}
	if err := r.NoStatus().Unmarshal(&rg).End(); err != nil {
		return ResultGroup{}, err
	}
	out := ResultGroup{Success: rg.Success}
	out.Results = make([]*Result, len(rg.Results))
	for i := range rg.Results {
		out.Results[i] = &Result{
			resource: &rg.Results[i].Resource,
			message:  rg.Results[i].Message,
			success:  rg.Results[i].Success,
			state:    rg.Results[i].State,
			status:   rg.Results[i].status,
		}
	}
	return out, nil
}

// SetGeneric enables or disables returning GenericResource objects within
// the results of the Apply function. If true all Resource objects embedded
// in Apply results will be GenericResource instances.
func (opts *ApplyOptions) SetGeneric(b bool) *ApplyOptions {
	if b {
		opts.unmarshal = applyGenericUnmarshal
	} else {
		opts.unmarshal = nil
	}
	return opts
}

// Generic returns true if the apply options are set to return
// GenericResource instances in Apply results.
func (opts *ApplyOptions) Generic() bool {
	return opts.unmarshal != nil
}
