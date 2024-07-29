//go:build (octopus || pacific || quincy) && ceph_preview

package striper

// #cgo LDFLAGS: -lrados -lradosstriper
// #include <stdlib.h>
// #include <radosstriper/libradosstriper.h>
import "C"

import (
	"unsafe"
)

// Stat returns metadata describing the striped object.
// This version of Stat uses an older API that does not provide time
// granularity below a second: the Nsec value of the StatInfo.ModTime field
// will always be zero.
//
// Implements:
//
//	int rados_striper_stat(rados_striper_t striper,
//	                       const char* soid,
//	                       uint64_t *psize,
//	                       time_t *pmtime);
func (s *Striper) Stat(soid string) (StatInfo, error) {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	var (
		size  C.uint64_t
		mtime C.time_t
	)
	ret := C.rados_striper_stat(
		s.striper,
		csoid,
		&size,
		&mtime)

	if ret < 0 {
		return StatInfo{}, getError(ret)
	}
	modts := Timespec{Sec: int64(mtime)}
	return StatInfo{
		Size:    uint64(size),
		ModTime: modts,
	}, nil
}
