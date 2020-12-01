package rados

// #cgo LDFLAGS: -lrados
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
//
import "C"

import (
	"unsafe"

	ts "github.com/ceph/go-ceph/internal/timespec"
)

// Timespec is a public type for the internal C 'struct timespec'
type Timespec ts.Timespec

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
	w.op = nil
	w.free()
}

func (w WriteOp) operate2(
	ioctx *IOContext, oid string, mtime *Timespec, flags OperationFlags) error {

	if err := ioctx.validate(); err != nil {
		return err
	}

	cOid := C.CString(oid)
	defer C.free(unsafe.Pointer(cOid))
	var cMtime *C.struct_timespec
	if mtime != nil {
		cMtime = &C.struct_timespec{}
		ts.CopyToCStruct(
			ts.Timespec(*mtime),
			ts.CTimespecPtr(cMtime))
	}

	ret := C.rados_write_op_operate2(
		w.op, ioctx.ioctx, cOid, cMtime, C.int(flags))
	return w.update(writeOp, ret)
}

// Operate will perform the operation(s).
func (w *WriteOp) Operate(ioctx *IOContext, oid string, flags OperationFlags) error {
	return w.operate2(ioctx, oid, nil, flags)
}

// OperateWithMtime will perform the operation while setting the modification
// time stamp to the supplied value.
func (w *WriteOp) OperateWithMtime(
	ioctx *IOContext, oid string, mtime Timespec, flags OperationFlags) error {

	return w.operate2(ioctx, oid, &mtime, flags)
}

func (w *WriteOp) operateCompat(ioctx *IOContext, oid string) error {
	switch err := w.Operate(ioctx, oid, OperationNoFlag).(type) {
	case nil:
		return nil
	case OperationError:
		return err.OpError
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

//  SetOmap appends the map `pairs` to the omap `oid`.
func (w *WriteOp) SetOmap(pairs map[string][]byte) {
	sos := newSetOmapStep(pairs)
	w.steps = append(w.steps, sos)
	C.rados_write_op_omap_set(
		w.op,
		sos.cKeys,
		sos.cValues,
		sos.cLengths,
		sos.cNum)
}
