package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
//
import "C"

import (
	"runtime"
	"unsafe"
)

type omapSetElement struct {
	// inputs:
	pairs map[string][]byte

	// arguments:
	cKeys    **C.char
	cValues  **C.char
	cLengths *C.size_t
	cNum     C.size_t

	// tracking vars:
	// strMem is a little bit hacky but we can no longer use defer
	// right after calling CString, since the memory needs to be
	// tied to the lifecycle of the WriteOp. For now just keep
	// an extra slice pointing to the C memory for easy cleanup
	// later.
	strMem []unsafe.Pointer
}

func newOmapSetElement(pairs map[string][]byte) *omapSetElement {
	var s C.size_t
	var c *C.char
	ptrSize := unsafe.Sizeof(c)

	c_keys := C.malloc(C.size_t(len(pairs)) * C.size_t(ptrSize))
	c_values := C.malloc(C.size_t(len(pairs)) * C.size_t(ptrSize))
	c_lengths := C.malloc(C.size_t(len(pairs)) * C.size_t(unsafe.Sizeof(s)))
	strMem := make([]unsafe.Pointer, 0)

	i := 0
	for key, value := range pairs {
		// key
		c_key_ptr := (**C.char)(unsafe.Pointer(uintptr(c_keys) + uintptr(i)*ptrSize))
		*c_key_ptr = C.CString(key)
		strMem = append(strMem, unsafe.Pointer(*c_key_ptr))

		// value and its length
		c_value_ptr := (**C.char)(unsafe.Pointer(uintptr(c_values) + uintptr(i)*ptrSize))

		var c_length C.size_t
		if len(value) > 0 {
			*c_value_ptr = (*C.char)(unsafe.Pointer(&value[0]))
			c_length = C.size_t(len(value))
		} else {
			*c_value_ptr = nil
			c_length = C.size_t(0)
		}

		c_length_ptr := (*C.size_t)(unsafe.Pointer(uintptr(c_lengths) + uintptr(i)*ptrSize))
		*c_length_ptr = c_length

		i++
	}

	oe := &omapSetElement{
		pairs:    pairs,
		cKeys:    (**C.char)(c_keys),
		cValues:  (**C.char)(c_values),
		cLengths: (*C.size_t)(c_lengths),
		cNum:     C.size_t(len(pairs)),
		strMem:   strMem,
	}
	runtime.SetFinalizer(oe, freeElement)
	return oe
}

func (oe *omapSetElement) free() {
	C.free(unsafe.Pointer(oe.cKeys))
	oe.cKeys = nil
	C.free(unsafe.Pointer(oe.cValues))
	oe.cValues = nil
	C.free(unsafe.Pointer(oe.cLengths))
	oe.cLengths = nil
	for _, p := range oe.strMem {
		C.free(p)
	}
	oe.strMem = nil
}

func (*omapSetElement) reset() {
}

func (*omapSetElement) update() error {
	return nil
}

// SetOmap appends the map `pairs` to the omap `oid`
func (ioctx *IOContext) SetOmap(oid string, pairs map[string][]byte) error {
	op := CreateWriteOp()
	defer op.Release()
	op.SetOmap(pairs)
	return op.operateCompat(ioctx, oid)
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

type omapRmKeysElement struct {
	// inputs:
	keys []string

	// arguments:
	cKeys **C.char
	cNum  C.size_t

	// tracking vars:
	strMem []unsafe.Pointer
}

func newOmapRmKeysElement(keys []string) *omapRmKeysElement {
	strMem := make([]unsafe.Pointer, 0)
	var c *C.char
	ptrSize := unsafe.Sizeof(c)

	c_keys := C.malloc(C.size_t(len(keys)) * C.size_t(ptrSize))

	i := 0
	for _, key := range keys {
		c_key_ptr := (**C.char)(unsafe.Pointer(uintptr(c_keys) + uintptr(i)*ptrSize))
		*c_key_ptr = C.CString(key)
		strMem = append(strMem, unsafe.Pointer(*c_key_ptr))
		i++
	}

	oe := &omapRmKeysElement{
		keys:   keys,
		cKeys:  (**C.char)(c_keys),
		cNum:   C.size_t(len(keys)),
		strMem: strMem,
	}
	runtime.SetFinalizer(oe, freeElement)
	return oe
}

func (oe *omapRmKeysElement) free() {
	C.free(unsafe.Pointer(oe.cKeys))
	oe.cKeys = nil
	for _, p := range oe.strMem {
		C.free(p)
	}
	oe.strMem = nil
}

func (*omapRmKeysElement) reset() {
}

func (*omapRmKeysElement) update() error {
	return nil
}

// RmOmapKeys removes the specified `keys` from the omap `oid`
func (ioctx *IOContext) RmOmapKeys(oid string, keys []string) error {
	op := CreateWriteOp()
	defer op.Release()
	op.RmOmapKeys(keys)
	return op.operateCompat(ioctx, oid)
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
