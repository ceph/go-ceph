// +build !luminous,!mimic
//
// Ceph Nautilus is the first release that includes rbd_namespace_create(),
// rbd_namespace_remove(), rbd_namespace_exists().

package rbd

// #cgo LDFLAGS: -lrbd
// #include <rados/librados.h>
// #include <rbd/librbd.h>
// #include <stdlib.h>
// #include <errno.h>
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/rados"
)

// NamespaceCreate creates the namespace for a given Rados IOContext.
//
// Implements:
//  int rbd_namespace_create(rados_ioctx_t io, const char *namespace_name);
func NamespaceCreate(ioctx *rados.IOContext, namespaceName string) error {
	if ioctx == nil {
		return ErrNoIOContext
	}
	if namespaceName == "" {
		return ErrNoNamespaceName
	}
	cNamespaceName := C.CString(namespaceName)
	defer C.free(unsafe.Pointer(cNamespaceName))

	ret := C.rbd_namespace_create(cephIoctx(ioctx), cNamespaceName)
	return getError(ret)
}

// NamespaceRemove removes a given namespace.
//
// Implements:
//  int rbd_namespace_remove(rados_ioctx_t io, const char *namespace_name);
func NamespaceRemove(ioctx *rados.IOContext, namespaceName string) error {
	if ioctx == nil {
		return ErrNoIOContext
	}
	if namespaceName == "" {
		return ErrNoNamespaceName
	}
	cNamespaceName := C.CString(namespaceName)
	defer C.free(unsafe.Pointer(cNamespaceName))

	ret := C.rbd_namespace_remove(cephIoctx(ioctx), cNamespaceName)
	return getError(ret)
}

// NamespaceExists checks whether a given namespace exists or not.
//
// Implements:
//  int rbd_namespace_exists(rados_ioctx_t io, const char *namespace_name, bool *exists);
func NamespaceExists(ioctx *rados.IOContext, namespaceName string) (bool, error) {
	if ioctx == nil {
		return false, ErrNoIOContext
	}
	if namespaceName == "" {
		return false, ErrNoNamespaceName
	}
	cNamespaceName := C.CString(namespaceName)
	defer C.free(unsafe.Pointer(cNamespaceName))

	var exists C.bool
	ret := C.rbd_namespace_exists(cephIoctx(ioctx), cNamespaceName, &exists)
	return bool(exists), getErrorIfNegative(ret)
}
