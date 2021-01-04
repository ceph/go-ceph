package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/cutil"
)

// SetOmap appends the map `pairs` to the omap `oid`
func (ioctx *IOContext) SetOmap(oid string, pairs map[string][]byte) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	numOfPairs := len(pairs)

	cKeys := cutil.NewCPtrSlice(numOfPairs)
	cValues := cutil.NewCPtrSlice(numOfPairs)
	cLengths := cutil.NewCSizeSlice(numOfPairs)
	defer cKeys.Free()
	defer cValues.Free()
	defer cLengths.Free()

	i := 0
	for key, value := range pairs {
		// key
		cKeys[i] = cutil.CPtr(C.CString(key))
		defer C.free(unsafe.Pointer(cKeys[i]))

		// value and its length
		if len(value) > 0 {
			cValues[i] = cutil.CPtr(C.CBytes(value))
			defer C.free(unsafe.Pointer(cValues[i]))
			cLengths[i] = cutil.CSize(len(value))
		} else {
			cValues[i] = nil
			cLengths[i] = 0
		}
		i++
	}

	op := C.rados_create_write_op()
	C.rados_write_op_omap_set(
		op,
		(**C.char)(cKeys.Ptr()),
		(**C.char)(cValues.Ptr()),
		(*C.size_t)(cLengths.Ptr()),
		C.size_t(numOfPairs))

	ret := C.rados_write_op_operate(op, ioctx.ioctx, c_oid, nil, 0)
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
	c_oid := C.CString(oid)
	c_start_after := C.CString(startAfter)
	c_filter_prefix := C.CString(filterPrefix)
	c_max_return := C.uint64_t(maxReturn)

	defer C.free(unsafe.Pointer(c_oid))
	defer C.free(unsafe.Pointer(c_start_after))
	defer C.free(unsafe.Pointer(c_filter_prefix))

	op := C.rados_create_read_op()

	var c_iter C.rados_omap_iter_t
	var c_prval C.int
	C.rados_read_op_omap_get_vals2(
		op,
		c_start_after,
		c_filter_prefix,
		c_max_return,
		&c_iter,
		nil,
		&c_prval,
	)

	ret := C.rados_read_op_operate(op, ioctx.ioctx, c_oid, 0)

	if int(ret) != 0 {
		return getError(ret)
	} else if int(c_prval) != 0 {
		return getError(c_prval)
	}

	for {
		var c_key *C.char
		var c_val *C.char
		var c_len C.size_t

		ret = C.rados_omap_get_next(c_iter, &c_key, &c_val, &c_len)

		if int(ret) != 0 {
			return getError(ret)
		}

		if c_key == nil {
			break
		}

		listFn(C.GoString(c_key), C.GoBytes(unsafe.Pointer(c_val), C.int(c_len)))
	}

	C.rados_omap_get_end(c_iter)
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
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	numOfKeys := len(keys)

	cKeys := cutil.NewCPtrSlice(numOfKeys)
	defer cKeys.Free()

	for i, key := range keys {
		cKeys[i] = cutil.CPtr(C.CString(key))
		defer C.free(unsafe.Pointer(cKeys[i]))
	}

	op := C.rados_create_write_op()
	C.rados_write_op_omap_rm_keys(
		op,
		(**C.char)(cKeys.Ptr()),
		C.size_t(numOfKeys))

	ret := C.rados_write_op_operate(op, ioctx.ioctx, c_oid, nil, 0)
	C.rados_release_write_op(op)

	return getError(ret)
}

// CleanOmap clears the omap `oid`
func (ioctx *IOContext) CleanOmap(oid string) error {
	c_oid := C.CString(oid)
	defer C.free(unsafe.Pointer(c_oid))

	op := C.rados_create_write_op()
	C.rados_write_op_omap_clear(op)

	ret := C.rados_write_op_operate(op, ioctx.ioctx, c_oid, nil, 0)
	C.rados_release_write_op(op)

	return getError(ret)
}
