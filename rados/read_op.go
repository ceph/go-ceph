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
	r.freeElements()
}

// Operate will perform the operation(s).
func (r *ReadOp) Operate(ioctx *IOContext, oid string, flags OperationFlags) error {
	if err := ioctx.validate(); err != nil {
		return err
	}

	cOid := C.CString(oid)
	defer C.free(unsafe.Pointer(cOid))

	r.reset()
	ret := C.rados_read_op_operate(r.op, ioctx.ioctx, cOid, C.int(flags))
	return r.finish("read", ret)
}

func (r *ReadOp) operateCompat(ioctx *IOContext, oid string) error {
	switch err := r.Operate(ioctx, oid, OperationNoFlag).(type) {
	case nil:
		return nil
	case OperationError:
		return err.Unwrap()
	default:
		return err
	}
}

// AssertExists assures the object targeted by the read op exists.
//
// Implements:
//  void rados_read_op_assert_exists(rados_read_op_t read_op);
func (r *ReadOp) AssertExists() {
	C.rados_read_op_assert_exists(r.op)
}

// GetOmapValues is used to iterate over a set, or sub-set, of omap keys
// as part of a read operation. An OmapGetElement is returned from this
// function. The OmapGetElement may be used to iterate over the key-value
// pairs after the Operate call has been performed.
func (r *ReadOp) GetOmapValues(startAfter, filterPrefix string, maxReturn uint64) *OmapGetElement {
	oge := newOmapGetElement(startAfter, filterPrefix, maxReturn)
	r.elements = append(r.elements, oge)
	C.rados_read_op_omap_get_vals2(
		r.op,
		oge.cStartAfter,
		oge.cFilterPrefix,
		C.uint64_t(oge.maxReturn),
		&oge.iter,
		&oge.more,
		&oge.rval,
	)
	return oge
}
