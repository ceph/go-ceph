package striper

// #cgo LDFLAGS: -lrados -lradosstriper
// #include <stdlib.h>
// #include <radosstriper/libradosstriper.h>
import "C"

import "unsafe"

// Write bytes from data into the striped object at the specified offset.
//
// Implements:
//
//	int rados_striper_write(rados_striper_t striper,
//	                        const char *soid,
//	                        const char *buf,
//	                        size_t len,
//	                        uint64_t off);
func (s *Striper) Write(soid string, data []byte, offset uint64) error {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	bufptr := (*C.char)(unsafe.Pointer(&data[0]))
	ret := C.rados_striper_write(
		s.striper,
		csoid,
		bufptr,
		C.size_t(len(data)),
		C.uint64_t(offset))
	return getError(ret)
}

// WriteFull writes all of the bytes in data to the striped object, truncating
// the object to the length of data.
//
// Implements:
//
//	int rados_striper_write_full(rados_striper_t striper,
//	                             const char *soid,
//	                             const char *buf,
//	                             size_t len);
func (s *Striper) WriteFull(soid string, data []byte) error {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	bufptr := (*C.char)(unsafe.Pointer(&data[0]))
	ret := C.rados_striper_write_full(
		s.striper,
		csoid,
		bufptr,
		C.size_t(len(data)))
	return getError(ret)
}

// Append the bytes in data to the end of the striped object.
//
// Implements:
//
//	int rados_striper_append(rados_striper_t striper,
//	                         const char *soid,
//	                         const char *buf,
//	                         size_t len);
func (s *Striper) Append(soid string, data []byte) error {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	bufptr := (*C.char)(unsafe.Pointer(&data[0]))
	ret := C.rados_striper_append(
		s.striper,
		csoid,
		bufptr,
		C.size_t(len(data)))
	return getError(ret)
}

// Remove a striped RADOS object.
//
// Implements:
//
//	int rados_striper_remove(rados_striper_t striper,
//	                         const char *soid);
func (s *Striper) Remove(soid string) error {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	ret := C.rados_striper_remove(s.striper, csoid)
	return getError(ret)
}

// Truncate a striped object, setting it to the specified size.
//
// Implements:
//
//	int rados_striper_trunc(rados_striper_t striper, const char *soid, uint64_t size);
func (s *Striper) Truncate(soid string, size uint64) error {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	ret := C.rados_striper_trunc(
		s.striper,
		csoid,
		C.uint64_t(size))
	return getError(ret)
}
