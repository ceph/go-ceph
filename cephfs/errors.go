package cephfs

/*
#include <errno.h>
*/
import "C"

import (
	"fmt"

	"github.com/ceph/go-ceph/internal/errutil"
)

// revive:disable:exported Temporarily live with stuttering

// CephFSError represents an error condition returned from the CephFS APIs.
type CephFSError int

// revive:enable:exported

// Error returns the error string for the CephFSError type.
func (e CephFSError) Error() string {
	errno, s := errutil.FormatErrno(int(e))
	if s == "" {
		return fmt.Sprintf("cephfs: ret=%d", errno)
	}
	return fmt.Sprintf("cephfs: ret=%d, %s", errno, s)
}

func getError(e C.int) error {
	if e == 0 {
		return nil
	}
	return CephFSError(e)
}

// Public go errors:

const (
	// ErrNotConnected may be returned when client is not connected
	// to a cluster.
	ErrNotConnected = CephFSError(-C.ENOTCONN)
)

// Private errors:

const (
	errNameTooLong = CephFSError(-C.ENAMETOOLONG)

	errInvalid = CephFSError(-C.EINVAL)
)
