package rbd

/*
#cgo LDFLAGS: -lrbd
#undef _GNU_SOURCE
#include <errno.h>
#include <stdlib.h>
#include <rbd/librbd.h>

extern int readIterateCallback(uint64_t, size_t, char*, uintptr_t);

// inline wrapper to cast uintptr_t to void*
static inline int wrap_rbd_read_iterate2(
	rbd_image_t image, uint64_t ofs, uint64_t len, uintptr_t arg) {
		return rbd_read_iterate2(image, ofs, len,
			(void*)readIterateCallback, (void*)arg);
};
*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/callbacks"
)

var readIterateCallbacks = callbacks.New()

// ReadIterateCallback defines the function signature needed for the
// ReadIterate callback function.
//
// The function will be called with the arguments: offset, length, buffer, and
// data. The offset and length correspond to a region of the image.  The buffer
// will contain the data read from the image, or be nil if the region in the
// image is zeroed (a hole).  The data value is an extra private parameter that
// can be set in the ReadIterateConfig and is meant to be used for passing
// arbitrary user-defined items to the callback function.
type ReadIterateCallback func(uint64, uint64, []byte, interface{}) int

// ReadIterateConfig is used to define the parameters of a ReadIterate call.
// Callback, Offset, and Length should always be specified when passed to
// ReadIterate. The data value is optional.
type ReadIterateConfig struct {
	Offset   uint64
	Length   uint64
	Callback ReadIterateCallback
	Data     interface{}
}

// ReadIterate walks over regions in the image, calling the callback function
// for each region. Typically, the size of each region is the stripe size of
// the image.
//
// Implements:
//  int rbd_read_iterate2(rbd_image_t image,
//                        uint64_t ofs,
//                        uint64_t len,
//                        int (*cb)(uint64_t, size_t, const char *, void *),
//                        void *arg);
func (image *Image) ReadIterate(config ReadIterateConfig) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}
	// the provided callback must be a real function
	if config.Callback == nil {
		return rbdError(C.EINVAL)
	}

	cbIndex := readIterateCallbacks.Add(config)
	defer readIterateCallbacks.Remove(cbIndex)

	ret := C.wrap_rbd_read_iterate2(
		image.image,
		C.uint64_t(config.Offset),
		C.uint64_t(config.Length),
		C.uintptr_t(cbIndex))

	return getError(ret)
}

//export readIterateCallback
func readIterateCallback(
	offset C.uint64_t, length C.size_t, cbuf *C.char, index uintptr) C.int {

	var gbuf []byte
	v := readIterateCallbacks.Lookup(index)
	config := v.(ReadIterateConfig)
	if cbuf != nil {
		// should we assert than length is < max val C.int?
		gbuf = C.GoBytes(unsafe.Pointer(cbuf), C.int(length))
	}
	return C.int(config.Callback(
		uint64(offset), uint64(length), gbuf, config.Data))
}
