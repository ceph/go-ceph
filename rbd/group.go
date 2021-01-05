package rbd

/*
#cgo LDFLAGS: -lrbd
#include <stdlib.h>
#include <rbd/librbd.h>
*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/rados"
)

// GroupCreate is used to create an image group.
//
// Implements:
//  int rbd_group_create(rados_ioctx_t p, const char *name);
func GroupCreate(ioctx *rados.IOContext, name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.rbd_group_create(cephIoctx(ioctx), cName)
	return getError(ret)
}

// GroupRemove is used to remove an image group.
//
// Implements:
//  int rbd_group_remove(rados_ioctx_t p, const char *name);
func GroupRemove(ioctx *rados.IOContext, name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.rbd_group_remove(cephIoctx(ioctx), cName)
	return getError(ret)
}

// GroupRename will rename an existing image group.
//
// Implements:
//  int rbd_group_rename(rados_ioctx_t p, const char *src_name,
//                       const char *dest_name);
func GroupRename(ioctx *rados.IOContext, src, dest string) error {
	cSrc := C.CString(src)
	defer C.free(unsafe.Pointer(cSrc))
	cDest := C.CString(dest)
	defer C.free(unsafe.Pointer(cDest))

	ret := C.rbd_group_rename(cephIoctx(ioctx), cSrc, cDest)
	return getError(ret)
}
