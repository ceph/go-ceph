//go:build ceph_preview

package rbd

/*
#cgo LDFLAGS: -lrbd
#undef _GNU_SOURCE
#include <errno.h>
#include <stdlib.h>
#include <rbd/librbd.h>

extern int diffIterateCallback(uint64_t, size_t, int, uintptr_t);

// rbd_diff_iterate3_fn matches the rbd_diff_iterate3 function signature.
typedef int(*rbd_diff_iterate3_fn)(rbd_image_t image, uint64_t from_snap_id,
	uint64_t ofs, uint64_t len, uint8_t include_parent, uint8_t whole_object,
	int (*cb)(uint64_t, size_t, int, void *), void *arg);

// rbd_diff_iterate3_dlsym take *fn as rbd_diff_iterate3_fn and calls the dynamically loaded
// rbd_diff_iterate3 function passed as 1st argument.
static inline int rbd_diff_iterate3_dlsym(void *fn, rbd_image_t image,
	uint64_t from_snap_id, uint64_t ofs, uint64_t len, uint8_t include_parent,
	uint8_t whole_object, uintptr_t arg) {
	// cast function pointer fn to rbd_diff_iterate3 and call the function
	return ((rbd_diff_iterate3_fn) fn)(image, from_snap_id, ofs, len, include_parent,
		whole_object, (void*)diffIterateCallback, (void*)arg);
}
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ceph/go-ceph/internal/dlsym"
)

var (
	diffIterateByIDOnce sync.Once
	diffIterateById     unsafe.Pointer
	diffIterateByIdErr  error
)

// DiffIterateByIDConfig is used to define the parameters of a DiffIterateByID call.
// Callback, Offset, and Length should always be specified when passed to
// DiffIterateByID. The other values are optional.
type DiffIterateByIDConfig struct {
	FromSnapID    uint64
	Offset        uint64
	Length        uint64
	IncludeParent DiffIncludeParent
	WholeObject   DiffWholeObject
	Callback      DiffIterateCallback
	Data          interface{}
}

// DiffIterateByID calls a callback on changed extents of an image.
//
// Calling DiffIterateByID will cause the callback specified in the
// DiffIterateConfig to be called as many times as there are changed
// regions in the image (controlled by the parameters as passed to librbd).
//
// See the documentation of DiffIterateCallback for a description of the
// arguments to the callback and the return behavior.
//
// Implements:
//
//	int rbd_diff_iterate3(rbd_image_t image,
//	                      uint64_t from_snap_id,
//	                      uint64_t ofs, uint64_t len,
//	                      uint8_t include_parent, uint8_t whole_object,
//	                      int (*cb)(uint64_t, size_t, int, void *),
//	                      void *arg);
func (image *Image) DiffIterateByID(config DiffIterateByIDConfig) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}
	if config.Callback == nil {
		return getError(C.EINVAL)
	}

	diffIterateByIDOnce.Do(func() {
		diffIterateById, diffIterateByIdErr = dlsym.LookupSymbol("rbd_diff_iterate3")
	})

	if diffIterateByIdErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, diffIterateByIdErr)
	}

	cbIndex := diffIterateCallbacks.Add(config)
	defer diffIterateCallbacks.Remove(cbIndex)

	ret := C.rbd_diff_iterate3_dlsym(
		diffIterateById,
		image.image,
		C.uint64_t(config.FromSnapID),
		C.uint64_t(config.Offset),
		C.uint64_t(config.Length),
		C.uint8_t(config.IncludeParent),
		C.uint8_t(config.WholeObject),
		C.uintptr_t(cbIndex))

	return getError(ret)
}
