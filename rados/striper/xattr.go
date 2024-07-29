//go:build ceph_preview

package striper

// #cgo LDFLAGS: -lrados -lradosstriper
// #include <stdlib.h>
// #include <radosstriper/libradosstriper.h>
import "C"

import (
	"errors"
	"unsafe"
)

var errStopIteration = errors.New("stop iteraton")

// GetXattr retrieves an extended attribute (xattr) of the given name from the
// specified striped object.
//
// Implements:
//
//	int rados_striper_getxattr(rados_striper_t striper,
//	                           const char *oid,
//	                           const char *name,
//	                           char *buf,
//	                           size_t len);
func (s *Striper) GetXattr(soid string, name string, data []byte) (int, error) {
	csoid := C.CString(soid)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(csoid))
	defer C.free(unsafe.Pointer(cName))

	ret := C.rados_striper_getxattr(
		s.striper,
		csoid,
		cName,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.size_t)(len(data)))

	if ret >= 0 {
		return int(ret), nil
	}
	return 0, getError(ret)
}

// SetXattr sets an extended attribute (xattr) of the given name on the
// specified striped object.
//
// Implements:
//
//	int rados_striper_setxattr(rados_striper_t striper,
//	                           const char *oid,
//	                           const char *name,
//	                           const char *buf,
//	                           size_t len);
func (s *Striper) SetXattr(soid string, name string, data []byte) error {
	csoid := C.CString(soid)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(csoid))
	defer C.free(unsafe.Pointer(cName))

	ret := C.rados_striper_setxattr(
		s.striper,
		csoid,
		cName,
		(*C.char)(unsafe.Pointer(&data[0])),
		(C.size_t)(len(data)))

	return getError(ret)
}

// RmXattr removes the extended attribute (xattr) of the given name from the
// striped object.
//
// Implements:
//
//	int rados_striper_rmxattr(rados_striper_t striper,
//	                          const char *oid,
//	                          const char *name);
func (s *Striper) RmXattr(soid string, name string) error {
	csoid := C.CString(soid)
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(csoid))
	defer C.free(unsafe.Pointer(cName))

	ret := C.rados_striper_rmxattr(s.striper, csoid, cName)
	return getError(ret)
}

// getXattrsNext wraps the function to fetch a xattr name/value pair
// from the iterator. If the iterator is exhausted the error value
// will be errStopIteration.
//
// Implements:
//
//	int rados_striper_getxattrs_next(rados_xattrs_iter_t iter,
//	                                 const char **name,
//	                                 const char **val,
//	                                 size_t *len);
func getXattrsNext(it C.rados_xattrs_iter_t) (string, []byte, error) {
	var (
		cName, cValue *C.char
		cLen          C.size_t
	)
	defer C.free(unsafe.Pointer(cName))
	defer C.free(unsafe.Pointer(cValue))

	ret := C.rados_striper_getxattrs_next(it, &cName, &cValue, &cLen)
	if ret < 0 {
		return "", nil, getError(ret)
	}
	if cName == nil && cValue == nil {
		return "", nil, errStopIteration
	}
	return C.GoString(cName), C.GoBytes(unsafe.Pointer(cValue), C.int(cLen)), nil
}

// ListXattrs returns a map containing all of the extended attributes (xattrs)
// for a striped object. The xattr names provide the key strings and the map's
// values are byte slices.
//
// Implements:
//
//	int rados_striper_getxattrs(rados_striper_t striper,
//	                            const char *oid,
//	                            rados_xattrs_iter_t *iter);
func (s *Striper) ListXattrs(soid string) (map[string][]byte, error) {
	csoid := C.CString(soid)
	defer C.free(unsafe.Pointer(csoid))

	var it C.rados_xattrs_iter_t
	ret := C.rados_striper_getxattrs(s.striper, csoid, &it)
	if ret < 0 {
		return nil, getError(ret)
	}
	defer C.rados_striper_getxattrs_end(it)

	m := make(map[string][]byte)
	for {
		xname, xvalue, err := getXattrsNext(it)
		if err == errStopIteration {
			break
		}
		if err != nil {
			return nil, err
		}
		m[xname] = xvalue
	}
	return m, nil
}
