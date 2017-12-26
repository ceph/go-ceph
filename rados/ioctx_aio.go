package rados

// #cgo LDFLAGS: -lrados
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import (
	"unsafe"
)

type Completion struct {
	cplt C.rados_completion_t
}

func NewCompletion() (*Completion, error) {
	completion := &Completion{}
	ret := C.rados_aio_create_completion(nil, nil, nil, &completion.cplt)
	if ret != 0 {
		return nil, GetRadosError(int(ret))
	}

	return completion, nil
}

func (completion *Completion) Release() {
	C.rados_aio_release(completion.cplt)
}

func (completion *Completion) WaitForComplete() {
	C.rados_aio_wait_for_complete(completion.cplt)
}

func (completion *Completion) WaitForSafe() {
	C.rados_aio_wait_for_safe(completion.cplt)
}

func (completion *Completion) IsComplete() bool {
	ret := C.rados_aio_is_complete(completion.cplt)
	if int(ret) == 1 {
		return true
	}

	return false
}

func (completion *Completion) IsSafe() bool {
	ret := C.rados_aio_is_safe(completion.cplt)
	if int(ret) == 1 {
		return true
	}

	return false
}

func (completion *Completion) ReturnValue() int {
	ret := C.rados_aio_get_return_value(completion.cplt)
	return int(ret)
}

func (ioctx *IOContext) AsyncRead(completion *Completion, oid string, data []byte, offset uint64) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}

	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	ret := C.rados_aio_read(
		ioctx.ioctx,
		c_oid,
		completion.cplt,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.size_t)(len(data)),
		(C.uint64_t)(offset))

	if ret >= 0 {
		return int(ret), nil
	} else {
		return 0, GetRadosError(int(ret))
	}
}

func (ioctx *IOContext) AsyncAppend(completion *Completion, oid string, data []byte) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	ret := C.rados_aio_append(ioctx.ioctx, c_oid,
		completion.cplt,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.size_t)(len(data)))
	return GetRadosError(int(ret))
}

func (ioctx *IOContext) AsyncWrite(completion *Completion, oid string, data []byte, offset uint64) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	ret := C.rados_aio_write(ioctx.ioctx, c_oid,
		completion.cplt,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.size_t)(len(data)),
		(C.uint64_t)(offset))

	return GetRadosError(int(ret))
}

func (ioctx *IOContext) AsyncDelete(completion *Completion, oid string) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	return GetRadosError(int(C.rados_aio_remove(ioctx.ioctx, c_oid, completion.cplt)))
}
