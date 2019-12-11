package rbd

// #cgo LDFLAGS: -lrbd
// #include <stdlib.h>
// #include <rbd/librbd.h>
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	// RBD image options.
	RbdImageOptionFormat            = C.RBD_IMAGE_OPTION_FORMAT
	RbdImageOptionFeatures          = C.RBD_IMAGE_OPTION_FEATURES
	RbdImageOptionOrder             = C.RBD_IMAGE_OPTION_ORDER
	RbdImageOptionStripeUnit        = C.RBD_IMAGE_OPTION_STRIPE_UNIT
	RbdImageOptionStripeCount       = C.RBD_IMAGE_OPTION_STRIPE_COUNT
	RbdImageOptionJournalOrder      = C.RBD_IMAGE_OPTION_JOURNAL_ORDER
	RbdImageOptionJournalSplayWidth = C.RBD_IMAGE_OPTION_JOURNAL_SPLAY_WIDTH
	RbdImageOptionJournalPool       = C.RBD_IMAGE_OPTION_JOURNAL_POOL
	RbdImageOptionFeaturesSet       = C.RBD_IMAGE_OPTION_FEATURES_SET
	RbdImageOptionFeaturesClear     = C.RBD_IMAGE_OPTION_FEATURES_CLEAR
	RbdImageOptionDataPool          = C.RBD_IMAGE_OPTION_DATA_POOL
	// introduced with Ceph Mimic
	//RbdImageOptionFlatten = C.RBD_IMAGE_OPTION_FLATTEN
)

type RbdImageOptions struct {
	options C.rbd_image_options_t
}

type RbdImageOption C.int

func NewRbdImageOptions() *RbdImageOptions {
	rio := &RbdImageOptions{}
	C.rbd_image_options_create(&rio.options)
	return rio
}

func (rio *RbdImageOptions) Destroy() {
	C.rbd_image_options_destroy(rio.options)
}

func (rio *RbdImageOptions) SetString(option RbdImageOption, value string) error {
	c_value := C.CString(value)
	defer C.free(unsafe.Pointer(c_value))

	ret := C.rbd_image_options_set_string(rio.options, C.int(option), c_value)
	if ret != 0 {
		return fmt.Errorf("%v, could not set option %v to \"%v\"",
			GetError(ret), option, value)
	}

	return nil
}

func (rio *RbdImageOptions) GetString(option RbdImageOption) (string, error) {
	value := make([]byte, 4096)

	ret := C.rbd_image_options_get_string(rio.options, C.int(option),
		(*C.char)(unsafe.Pointer(&value[0])),
		C.size_t(len(value)))
	if ret != 0 {
		return "", fmt.Errorf("%v, could not get option %v", GetError(ret), option)
	}

	return C.GoString((*C.char)(unsafe.Pointer(&value[0]))), nil
}

func (rio *RbdImageOptions) SetUint64(option RbdImageOption, value uint64) error {
	c_value := C.uint64_t(value)

	ret := C.rbd_image_options_set_uint64(rio.options, C.int(option), c_value)
	if ret != 0 {
		return fmt.Errorf("%v, could not set option %v to \"%v\"",
			GetError(ret), option, value)
	}

	return nil
}

func (rio *RbdImageOptions) GetUint64(option RbdImageOption) (uint64, error) {
	var c_value C.uint64_t

	ret := C.rbd_image_options_get_uint64(rio.options, C.int(option), &c_value)
	if ret != 0 {
		return 0, fmt.Errorf("%v, could not get option %v", GetError(ret), option)
	}

	return uint64(c_value), nil
}

func (rio *RbdImageOptions) IsSet(option RbdImageOption) (bool, error) {
	var c_set C.bool

	ret := C.rbd_image_options_is_set(rio.options, C.int(option), &c_set)
	if ret != 0 {
		return false, fmt.Errorf("%v, could not check option %v", GetError(ret), option)
	}

	return bool(c_set), nil
}

func (rio *RbdImageOptions) Unset(option RbdImageOption) error {
	ret := C.rbd_image_options_unset(rio.options, C.int(option))
	if ret != 0 {
		return fmt.Errorf("%v, could not unset option %v", GetError(ret), option)
	}

	return nil
}

func (rio *RbdImageOptions) Clear() {
	C.rbd_image_options_clear(rio.options)
}

func (rio *RbdImageOptions) IsEmpty() bool {
	ret := C.rbd_image_options_is_empty(rio.options)
	return ret != 0
}
