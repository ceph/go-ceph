package striper

/*
#include <errno.h>
*/
import "C"

import (
	"github.com/ceph/go-ceph/internal/errutil"
)

// errRange is returned by librados when a value does not fit in the supplied
// buffer (ERANGE).
var errRange = getError(-C.ERANGE)

// radosStriperError represents an error condition returned from the Ceph
// rados striper APIs.
type radosStriperError int

// Error returns the error string for the radosStriperError type.
func (e radosStriperError) Error() string {
	return errutil.FormatErrorCode("rados", int(e))
}

func (e radosStriperError) ErrorCode() int {
	return int(e)
}

func getError(e C.int) error {
	if e == 0 {
		return nil
	}
	return radosStriperError(e)
}
