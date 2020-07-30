package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import "unsafe"

// CreateSnap creates a pool-wide snapshot.
//
// Implements:
// int rados_ioctx_snap_create(rados_ioctx_t io, const char *snapname)
func (ioctx *IOContext) CreateSnap(snapName string) error {
	if err := ioctx.validate(); err != nil {
		return err
	}

	cSnapName := C.CString(snapName)
	defer C.free(unsafe.Pointer(cSnapName))

	ret := C.rados_ioctx_snap_create(ioctx.ioctx, cSnapName)
	return getError(ret)
}

// RemoveSnap deletes the pool snapshot.
//
// Implements:
//  int rados_ioctx_snap_remove(rados_ioctx_t io, const char *snapname)
func (ioctx *IOContext) RemoveSnap(snapName string) error {
	if err := ioctx.validate(); err != nil {
		return err
	}

	cSnapName := C.CString(snapName)
	defer C.free(unsafe.Pointer(cSnapName))

	ret := C.rados_ioctx_snap_remove(ioctx.ioctx, cSnapName)
	return getError(ret)
}
