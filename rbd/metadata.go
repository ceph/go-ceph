package rbd

// #cgo LDFLAGS: -lrbd
// #include <stdlib.h>
// #include <rbd/librbd.h>
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/retry"
)

// GetMetadata returns the metadata string associated with the given key.
//
// Implements:
//  int rbd_metadata_get(rbd_image_t image, const char *key, char *value, size_t *vallen)
func (image *Image) GetMetadata(key string) (string, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return "", err
	}

	c_key := C.CString(key)
	defer C.free(unsafe.Pointer(c_key))

	var (
		buf []byte
		err error
	)
	retry.WithSizes(4096, 262144, func(size int) retry.Hint {
		csize := C.size_t(size)
		buf = make([]byte, csize)
		// rbd_metadata_get is a bit quirky and *does not* update the size
		// value if the size passed in >= the needed size.
		ret := C.rbd_metadata_get(
			image.image, c_key, (*C.char)(unsafe.Pointer(&buf[0])), &csize)
		err = getError(ret)
		return retry.Size(int(csize)).If(err == errRange)
	})
	if err != nil {
		return "", err
	}
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), nil
}

// SetMetadata updates the metadata string associated with the given key.
//
// Implements:
//  int rbd_metadata_set(rbd_image_t image, const char *key, const char *value)
func (image *Image) SetMetadata(key string, value string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_key := C.CString(key)
	c_value := C.CString(value)
	defer C.free(unsafe.Pointer(c_key))
	defer C.free(unsafe.Pointer(c_value))

	ret := C.rbd_metadata_set(image.image, c_key, c_value)
	if ret < 0 {
		return rbdError(ret)
	}

	return nil
}

// RemoveMetadata clears the metadata associated with the given key.
//
// Implements:
//  int rbd_metadata_remove(rbd_image_t image, const char *key)
func (image *Image) RemoveMetadata(key string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_key := C.CString(key)
	defer C.free(unsafe.Pointer(c_key))

	ret := C.rbd_metadata_remove(image.image, c_key)
	if ret < 0 {
		return rbdError(ret)
	}

	return nil
}
