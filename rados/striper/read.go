//go:build ceph_preview

package striper

// #cgo LDFLAGS: -lrados -lradosstriper
// #include <stdlib.h>
// #include <radosstriper/libradosstriper.h>
import "C"

import (
	"unsafe"
)

// Read bytes into data from the striped object at the specified offset.
//
// Implements:
//
//	int rados_striper_read(rados_striper_t striper,
//	                       const char *soid,
//	                       const char *buf,
//	                       size_t len,
//	                       uint64_t off);
func (s *Striper) Read(soid string, data []byte, offset uint64) (int, error) {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	var bufptr *C.char
	if len(data) > 0 {
		bufptr = (*C.char)(unsafe.Pointer(&data[0]))
	}

	ret := C.rados_striper_read(
		s.striper,
		csoid,
		bufptr,
		C.size_t(len(data)),
		C.uint64_t(offset))
	if ret >= 0 {
		return int(ret), nil
	}
	return 0, getError(ret)
}
