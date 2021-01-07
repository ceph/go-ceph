package rgw

/*
#include <errno.h>
*/
import "C"

import (
	"errors"

	"github.com/ceph/go-ceph/internal/errutil"
)

// rgwError represents an error condition returned from the RGW APIs.
type rgwError int

// Error returns the error string for the rgwError type.
func (e rgwError) Error() string {
	return errutil.FormatErrorCode("rgw", int(e))
}

func (e rgwError) ErrorCode() int {
	return int(e)
}

func getError(e C.int) error {
	if e == 0 {
		return nil
	}
	return rgwError(e)
}

// getErrorIfNegative converts a ceph return code to error if negative.
// This is useful for functions that return a usable positive value on
// success but a negative error number on error.
func getErrorIfNegative(ret C.int) error {
	if ret >= 0 {
		return nil
	}
	return getError(ret)
}

// Public go errors:

var (
	// ErrEmptyArgument may be returned if a function argument is passed
	// a zero-length slice or map.
	ErrEmptyArgument = errors.New("Argument must contain at least one item")
)

// Public RGWErrors:

const (
	// ErrNotConnected may be returned when client is not connected
	// to a cluster.
	ErrNotConnected = rgwError(-C.ENOTCONN)
)

// Private errors:

const (
	errInvalid     = rgwError(-C.EINVAL)
	errNameTooLong = rgwError(-C.ENAMETOOLONG)
	errNoEntry     = rgwError(-C.ENOENT)
	errRange       = rgwError(-C.ERANGE)
)
