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

// WriteOp manages a set of discrete actions that will be performed together
// atomically.
type WriteOp struct {
	operation
	op C.rados_write_op_t
}

// CreateWriteOp returns a newly constructed write operation.
func CreateWriteOp() *WriteOp {
	return &WriteOp{
		op: C.rados_create_write_op(),
	}
}

// Release the resources associated with this write operation.
func (w *WriteOp) Release() {
	C.rados_release_write_op(w.op)
	w.freeElements()
}

// Operate will perform the operation(s).
func (w *WriteOp) Operate(ioctx *IOContext, oid string) error {
	cOid := C.CString(oid)
	defer C.free(unsafe.Pointer(cOid))

	w.reset()
	ret := C.rados_write_op_operate(w.op, ioctx.ioctx, cOid, nil, 0)
	return w.finish("write", ret)
}

func (w *WriteOp) operateCompat(ioctx *IOContext, oid string) error {
	switch err := w.Operate(ioctx, oid).(type) {
	case nil:
		return nil
	case OperationError:
		return err.Unwrap()
	default:
		return err
	}
}

// Create a rados object.
func (w *WriteOp) Create(exclusive CreateOption) {
	// category, the 3rd param, is deprecated and has no effect so we do not
	// implement it in go-ceph
	C.rados_write_op_create(w.op, C.int(exclusive), nil)
}
