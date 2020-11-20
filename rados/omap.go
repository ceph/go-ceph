package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
//
import "C"

import (
	"unsafe"
)

const ptrSize = unsafe.Sizeof(unsafe.Pointer(nil))

// SetOmap appends the map `pairs` to the omap `oid`
func (ioctx *IOContext) SetOmap(oid string, pairs map[string][]byte) error {
	cOid := C.CString(oid)
	defer C.free(unsafe.Pointer(cOid))

	numOfPairs := uintptr(len(pairs))

	cKeys := C.malloc(C.size_t(numOfPairs * ptrSize))
	cValues := C.malloc(C.size_t(numOfPairs * ptrSize))
	cLengths := C.malloc(C.size_t(numOfPairs * C.sizeof_size_t))

	defer C.free(unsafe.Pointer(cKeys))
	defer C.free(unsafe.Pointer(cValues))
	defer C.free(unsafe.Pointer(cLengths))

	var i uintptr
	for key, value := range pairs {
		// key
		cKeyPtr := (**C.char)(unsafe.Pointer(uintptr(cKeys) + i*ptrSize))
		*cKeyPtr = C.CString(key)
		defer C.free(unsafe.Pointer(*cKeyPtr))

		// value and its length
		cValuePtr := (**C.char)(unsafe.Pointer(uintptr(cValues) + i*ptrSize))

		var cLength C.size_t
		if len(value) > 0 {
			*cValuePtr = (*C.char)(unsafe.Pointer(&value[0]))
			cLength = C.size_t(len(value))
		} else {
			*cValuePtr = nil
			cLength = C.size_t(0)
		}

		cLengthPtr := (*C.size_t)(unsafe.Pointer(uintptr(cLengths) + i*ptrSize))
		*cLengthPtr = cLength

		i++
	}

	op := C.rados_create_write_op()
	C.rados_write_op_omap_set(
		op,
		(**C.char)(cKeys),
		(**C.char)(cValues),
		(*C.size_t)(cLengths),
		C.size_t(numOfPairs))

	ret := C.rados_write_op_operate(op, ioctx.ioctx, cOid, nil, 0)
	C.rados_release_write_op(op)

	return getError(ret)
}

// OmapListFunc is the type of the function called for each omap key
// visited by ListOmapValues
type OmapListFunc func(key string, value []byte)

// ListOmapValues iterates over the keys and values in an omap by way of
// a callback function.
//
// `startAfter`: iterate only on the keys after this specified one
// `filterPrefix`: iterate only on the keys beginning with this prefix
// `maxReturn`: iterate no more than `maxReturn` key/value pairs
// `listFn`: the function called at each iteration
func (ioctx *IOContext) ListOmapValues(oid string, startAfter string, filterPrefix string, maxReturn int64, listFn OmapListFunc) error {
	cOid := C.CString(oid)
	cStartAfter := C.CString(startAfter)
	cFilterPrefix := C.CString(filterPrefix)
	cMaxReturn := C.uint64_t(maxReturn)

	defer C.free(unsafe.Pointer(cOid))
	defer C.free(unsafe.Pointer(cStartAfter))
	defer C.free(unsafe.Pointer(cFilterPrefix))

	op := C.rados_create_read_op()

	var cIter C.rados_omap_iter_t
	var cPrval C.int
	C.rados_read_op_omap_get_vals2(
		op,
		cStartAfter,
		cFilterPrefix,
		cMaxReturn,
		&cIter,
		nil,
		&cPrval,
	)

	ret := C.rados_read_op_operate(op, ioctx.ioctx, cOid, 0)

	if int(ret) != 0 {
		return getError(ret)
	} else if int(cPrval) != 0 {
		return getError(cPrval)
	}

	for {
		var cKey *C.char
		var cVal *C.char
		var cLen C.size_t

		ret = C.rados_omap_get_next(cIter, &cKey, &cVal, &cLen)

		if int(ret) != 0 {
			return getError(ret)
		}

		if cKey == nil {
			break
		}

		listFn(C.GoString(cKey), C.GoBytes(unsafe.Pointer(cVal), C.int(cLen)))
	}

	C.rados_omap_get_end(cIter)
	C.rados_release_read_op(op)

	return nil
}

// GetOmapValues fetches a set of keys and their values from an omap and returns then as a map
// `startAfter`: retrieve only the keys after this specified one
// `filterPrefix`: retrieve only the keys beginning with this prefix
// `maxReturn`: retrieve no more than `maxReturn` key/value pairs
func (ioctx *IOContext) GetOmapValues(oid string, startAfter string, filterPrefix string, maxReturn int64) (map[string][]byte, error) {
	omap := map[string][]byte{}

	err := ioctx.ListOmapValues(
		oid, startAfter, filterPrefix, maxReturn,
		func(key string, value []byte) {
			omap[key] = value
		},
	)

	return omap, err
}

// GetAllOmapValues fetches all the keys and their values from an omap and returns then as a map
// `startAfter`: retrieve only the keys after this specified one
// `filterPrefix`: retrieve only the keys beginning with this prefix
// `iteratorSize`: internal number of keys to fetch during a read operation
func (ioctx *IOContext) GetAllOmapValues(oid string, startAfter string, filterPrefix string, iteratorSize int64) (map[string][]byte, error) {
	omap := map[string][]byte{}
	omapSize := 0

	for {
		err := ioctx.ListOmapValues(
			oid, startAfter, filterPrefix, iteratorSize,
			func(key string, value []byte) {
				omap[key] = value
				startAfter = key
			},
		)

		if err != nil {
			return omap, err
		}

		// End of omap
		if len(omap) == omapSize {
			break
		}

		omapSize = len(omap)
	}

	return omap, nil
}

// RmOmapKeys removes the specified `keys` from the omap `oid`
func (ioctx *IOContext) RmOmapKeys(oid string, keys []string) error {
	cOid := C.CString(oid)
	defer C.free(unsafe.Pointer(cOid))

	cKeys := C.malloc(C.size_t(uintptr(len(keys)) * ptrSize))
	defer C.free(unsafe.Pointer(cKeys))

	var i uintptr
	for _, key := range keys {
		cKeyPtr := (**C.char)(unsafe.Pointer(uintptr(cKeys) + i*ptrSize))
		*cKeyPtr = C.CString(key)
		defer C.free(unsafe.Pointer(*cKeyPtr))
		i++
	}

	op := C.rados_create_write_op()
	C.rados_write_op_omap_rm_keys(
		op,
		(**C.char)(cKeys),
		C.size_t(len(keys)))

	ret := C.rados_write_op_operate(op, ioctx.ioctx, cOid, nil, 0)
	C.rados_release_write_op(op)

	return getError(ret)
}

// CleanOmap clears the omap `oid`
func (ioctx *IOContext) CleanOmap(oid string) error {
	cOid := C.CString(oid)
	defer C.free(unsafe.Pointer(cOid))

	op := C.rados_create_write_op()
	C.rados_write_op_omap_clear(op)

	ret := C.rados_write_op_operate(op, ioctx.ioctx, cOid, nil, 0)
	C.rados_release_write_op(op)

	return getError(ret)
}
