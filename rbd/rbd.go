package rbd

// #cgo LDFLAGS: -lrbd
// #include <stdlib.h>
// #include <rados/librados.h>
// #include <rbd/librbd.h>
import "C"

import (
    "github.com/noahdesu/go-rados/rados"
    "fmt"
    "unsafe"
    "bytes"
)

//
type RBDError int

//
func (e RBDError) Error() string {
    return fmt.Sprintf("rbd: ret=%d", e)
}

//
func Version() (int, int, int) {
    var c_major, c_minor, c_patch C.int
    C.rbd_version(&c_major, &c_minor, &c_patch)
    return int(c_major), int(c_minor), int(c_patch)
}

// Create
func Create(ioctx *rados.IOContext, name string, size uint64) error {
    var c_order C.int
    c_name := C.CString(name)
    defer C.free(unsafe.Pointer(c_name))
    ret := C.rbd_create(C.rados_ioctx_t(ioctx.Pointer()), c_name, C.uint64_t(size), &c_order)
    if ret < 0 {
        return RBDError(ret)
    } else {
        return nil
    }
}

// GetImageNames returns the list of current RBD images.
func GetImageNames(ioctx *rados.IOContext) (names []string, err error) {
    buf := make([]byte, 4096)
    for {
        size := C.size_t(len(buf))
        ret := C.rbd_list(C.rados_ioctx_t(ioctx.Pointer()),
            (*C.char)(unsafe.Pointer(&buf[0])), &size)
        if ret == -34 { // FIXME
            buf = make([]byte, size)
            continue
        } else if ret < 0 {
            return nil, RBDError(ret)
        }
        tmp := bytes.Split(buf[:size-1], []byte{0})
        for _, s := range tmp {
            if len(s) > 0 {
                name := C.GoString((*C.char)(unsafe.Pointer(&s[0])))
                names = append(names, name)
            }
        }
        return names, nil
    }
}
