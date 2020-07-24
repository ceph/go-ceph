package cephfs

/*
#include <errno.h>
*/
import "C"

import (
	"errors"
	"fmt"

	"github.com/ceph/go-ceph/internal/errutil"
)

// cephFSError represents an error condition returned from the CephFS APIs.
type cephFSError int

// Error returns the error string for the cephFSError type.
func (e cephFSError) Error() string {
	errno, s := errutil.FormatErrno(int(e))
	if s == "" {
		return fmt.Sprintf("cephfs: ret=%d", errno)
	}
	return fmt.Sprintf("cephfs: ret=%d, %s", errno, s)
}

func (e cephFSError) Errno() int {
	return int(e)
}

func getError(e C.int) error {
	if e == 0 {
		return nil
	}
	return cephFSError(e)
}

// Public go errors:

var (
	// ErrEmptyArgument may be returned if a function argument is passed
	// a zero-length slice or map.
	ErrEmptyArgument = errors.New("Argument must contain at least one item")
)

// Public CephFSErrors:

const (
	// ErrNotConnected may be returned when client is not connected
	// to a cluster.
	ErrNotConnected = cephFSError(-C.ENOTCONN)
)

// Private errors:

const (
	errInvalid     = cephFSError(-C.EINVAL)
	errNameTooLong = cephFSError(-C.ENAMETOOLONG)
	errNoEntry     = cephFSError(-C.ENOENT)
)
