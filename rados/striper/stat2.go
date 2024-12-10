//go:build !(octopus || pacific || quincy)

package striper

// #cgo LDFLAGS: -lrados -lradosstriper
// #include <stdlib.h>
// #include <radosstriper/libradosstriper.h>
import "C"

import (
	"unsafe"

	ts "github.com/ceph/go-ceph/internal/timespec"
)

// Stat returns metadata describing the striped object.
//
// Implements:
//
//	int rados_striper_stat2(rados_striper_t striper,
//	                       const char* soid,
//	                       uint64_t *psize,
//	                       struct timespec *pmtime);
func (s *Striper) Stat(soid string) (StatInfo, error) {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	var (
		size  C.uint64_t
		mtime C.struct_timespec
	)
	ret := C.rados_striper_stat2(
		s.striper,
		csoid,
		&size,
		&mtime)

	if ret < 0 {
		return StatInfo{}, getError(ret)
	}
	return StatInfo{
		Size:    uint64(size),
		ModTime: Timespec(ts.CStructToTimespec(ts.CTimespecPtr(&mtime))),
	}, nil
}
