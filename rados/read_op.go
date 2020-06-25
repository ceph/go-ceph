package rados

// #cgo LDFLAGS: -lrados
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
//
import "C"

import (
	"unsafe"
)

// ReadOp manages a set of discrete object read actions that will be performed
// together atomically.
type ReadOp struct {
	operation
	op C.rados_read_op_t
}

// CreateReadOp returns a newly constructed read operation.
func CreateReadOp() *ReadOp {
	return &ReadOp{
		op: C.rados_create_read_op(),
	}
}

// Release the resources associated with this read operation.
func (r *ReadOp) Release() {
	C.rados_release_read_op(r.op)
	r.op = nil
	r.free()
}

// Operate will perform the operation(s).
func (r *ReadOp) Operate(ioctx *IOContext, oid string, flags OperationFlags) error {
	if err := ioctx.validate(); err != nil {
		return err
	}

	cOid := C.CString(oid)
	defer C.free(unsafe.Pointer(cOid))

	ret := C.rados_read_op_operate(r.op, ioctx.ioctx, cOid, C.int(flags))
	return r.update(readOp, ret)
}

func (r *ReadOp) operateCompat(ioctx *IOContext, oid string) error {
	switch err := r.Operate(ioctx, oid, OperationNoFlag).(type) {
	case nil:
		return nil
	case OperationError:
		return err.OpError
	default:
		return err
	}
}
