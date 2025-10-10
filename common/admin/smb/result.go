//go:build !(octopus || pacific || quincy || reef || squid)

package smb

import (
	"encoding/json"
	"fmt"
)

// resultCommon holds common fields in result JSON.
type resultCommon struct {
	Resource resourceEntry `json:"resource"`
	Message  string        `json:"msg"`
	Success  bool          `json:"success"`
	State    string        `json:"state"`
}

// Result represents the result of applying a new/changed resource.
type Result struct {
	resource Resource
	message  string
	success  bool

	state  string
	status map[string]any
}

// UnmarshalJSON support unmarshalling JSON to a Result.
func (r *Result) UnmarshalJSON(data []byte) error {
	rc := &resultCommon{}
	if err := json.Unmarshal(data, &rc); err != nil {
		return err
	}
	r.resource = rc.Resource.r
	r.message = rc.Message
	r.success = rc.Success
	r.state = rc.State
	r.status = map[string]any{}
	// stash extra stuff that may be in the result
	return json.Unmarshal(data, &r.status)
}

// Ok returns true if the resource modification was a success.
func (r *Result) Ok() bool {
	return r.success
}

// Resource returns the resource changed.
func (r *Result) Resource() Resource {
	return r.resource
}

// Message returns an optional string describing the modification state.
func (r *Result) Message() string {
	return r.message
}

// Error supports treating a failed result as a Go error.
func (r *Result) Error() string {
	if r.success {
		return ""
	}
	return fmt.Sprintf("%s: %s", r.resource.Identity(), r.message)
}

// State returns a short string describing the state of the resource.
func (r *Result) State() string {
	return r.state
}

// Dump additional fields returned with the result.
func (r *Result) Dump() map[string]any {
	return r.status
}

// ResultGroup contains a series of Results and summarizes if a modifcation was
// a success overall.
type ResultGroup struct {
	Success bool      `json:"success"`
	Results []*Result `json:"results"`
}

// Ok returns true if all the resource modifications were successful.
func (rgroup ResultGroup) Ok() bool {
	return rgroup.Success
}

// ErrorResults returns a slice of results containing items that were not
// successful.
func (rgroup ResultGroup) ErrorResults() []*Result {
	errs := []*Result{}
	for _, result := range rgroup.Results {
		if !result.Ok() {
			errs = append(errs, result)
		}
	}
	return errs
}

// Error supports treating a failed ResultGroup as a Go error.
func (rgroup ResultGroup) Error() string {
	errs := rgroup.ErrorResults()
	if len(errs) == 0 {
		return ""
	}
	ec := ""
	for i, err := range errs {
		sep := "; "
		if i == 0 {
			sep = ""
		}
		ec = fmt.Sprintf("%s%s%s", ec, sep, err.Error())
	}
	return fmt.Sprintf("%d resource errors: %s", len(errs), ec)
}
