package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import (
    "unsafe"
    "fmt"
)

type RadosError int

func (e RadosError) Error() string {
    return fmt.Sprintf("rados: ret=%d", e)
}

func Version() (int, int, int) {
    var c_major, c_minor, c_patch C.int
    C.rados_version(&c_major, &c_minor, &c_patch)
    return int(c_major), int(c_minor), int(c_patch)
}

func Open(id string) (*Conn, error) {
    c_id := C.CString(id)
    defer C.free(unsafe.Pointer(c_id))

    conn := &Conn{}
    ret := C.rados_create(&conn.cluster, c_id)

    if ret == 0 {
        return conn, nil
    } else {
        return nil, RadosError(int(ret))
    }
}
