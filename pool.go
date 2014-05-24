package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import "unsafe"

// Pool represents a context for performing I/O within a pool.
type Pool struct {
    ioctx C.rados_ioctx_t
}

// Write writes len(data) bytes to the object with key oid starting at byte
// offset offset. It returns an error, if any.
func (p *Pool) Write(oid string, data []byte, offset uint64) error {
    c_oid := C.CString(oid)
    defer C.free(unsafe.Pointer(c_oid))

    ret := C.rados_write(p.ioctx, c_oid,
        (*C.char)(unsafe.Pointer(&data[0])),
        (C.size_t)(len(data)),
        (C.uint64_t)(offset))

    if ret == 0 {
        return nil
    } else {
        return RadosError(int(ret))
    }
}

// Read reads up to len(data) bytes from the object with key oid starting at byte
// offset offset. It returns the number of bytes read and an error, if any.
func (p *Pool) Read(oid string, data []byte, offset uint64) (int, error) {
    if len(data) == 0 {
        return 0, nil
    }

    c_oid := C.CString(oid)
    defer C.free(unsafe.Pointer(c_oid))

    ret := C.rados_read(
        p.ioctx,
        c_oid,
        (*C.char)(unsafe.Pointer(&data[0])),
        (C.size_t)(len(data)),
        (C.uint64_t)(offset))

    if ret >= 0 {
        return int(ret), nil
    } else {
        return 0, RadosError(int(ret))
    }
}

// Delete deletes the object with key oid. It returns an error, if any.
func (p *Pool) Delete(oid string) error {
    c_oid := C.CString(oid)
    defer C.free(unsafe.Pointer(c_oid))

    ret := C.rados_remove(p.ioctx, c_oid)

    if ret == 0 {
        return nil
    } else {
        return RadosError(int(ret))
    }
}

// Truncate resizes the object with key oid to size size. If the operation
// enlarges the object, the new area is logically filled with zeroes. If the
// operation shrinks the object, the excess data is removed. It returns an
// error, if any.
func (p *Pool) Truncate(oid string, size uint64) error {
    c_oid := C.CString(oid)
    defer C.free(unsafe.Pointer(c_oid))

    ret := C.rados_trunc(p.ioctx, c_oid, (C.uint64_t)(size))

    if ret == 0 {
        return nil
    } else {
        return RadosError(int(ret))
    }
}
