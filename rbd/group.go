package rbd

/*
#cgo LDFLAGS: -lrbd
#include <stdlib.h>
#include <rbd/librbd.h>
*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/cutil"
	"github.com/ceph/go-ceph/internal/retry"
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

// GroupList returns a slice of image group names.
//
// Implements:
//  int rbd_group_list(rados_ioctx_t p, char *names, size_t *size);
func GroupList(ioctx *rados.IOContext) ([]string, error) {
	var (
		buf []byte
		err error
		ret C.int
	)
	retry.WithSizes(1024, 262144, func(size int) retry.Hint {
		cSize := C.size_t(size)
		buf = make([]byte, cSize)
		ret = C.rbd_group_list(
			cephIoctx(ioctx),
			(*C.char)(unsafe.Pointer(&buf[0])),
			&cSize)
		err = getErrorIfNegative(ret)
		return retry.Size(int(cSize)).If(err == errRange)
	})

	if err != nil {
		return nil, err
	}

	// cSize is not set to the expected size when it is sufficiently large
	// but ret will be set to the size in a non-error condition.
	groups := cutil.SplitBuffer(buf[:ret])
	return groups, nil
}
