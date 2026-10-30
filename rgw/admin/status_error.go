//go:build ceph_preview

package admin

import "fmt"

// StatusError reports a failed admin API request, exposing StatusCode, the HTTP
// status code of the response. Retrieve it with errors.As.
type StatusError struct {
	StatusCode int
	code       string
}

// Error implements the error interface.
func (e StatusError) Error() string {
	return fmt.Sprintf("%s (HTTP status %d)", e.code, e.StatusCode)
}

// Is reports whether the error matches one of the exported error reasons
// (e.g. ErrNoSuchUser).
func (e StatusError) Is(target error) bool { return target == errorReason(e.code) }

// As lets errors.As retrieve a StatusError from the error returned by the admin
// API, without exposing the internal statusError type.
func (e statusError) As(target any) bool {
	if se, ok := target.(*StatusError); ok {
		se.StatusCode = e.statusCode
		se.code = e.Code
		return true
	}
	return false
}
