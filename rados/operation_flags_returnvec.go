//go:build ceph_preview

package rados

// #cgo LDFLAGS: -lrados
// #include <rados/librados.h>
//
import "C"

const (
	// OperationReturnVec asks the OSD to return the result and output data
	// of each action even when the operation writes, which it otherwise
	// discards.
	OperationReturnVec = OperationFlags(C.LIBRADOS_OPERATION_RETURNVEC)
)
