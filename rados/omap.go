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

// omapSetElement is a write op element used to track state, especially
// C memory, across the setup and use of a WriteOp.
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

// OmapKeyValue items are returned by the OmapGetElement's Next call.
type OmapKeyValue struct {
	Key   string
	Value []byte
}

// OmapGetElement values are used to get the results of an GetOmapValues call
// on a WriteOp. Until the Operate method of the WriteOp is called the Next
// call will return an error. After Operate is called, the Next call will
// return valid results.
// The life cycle of the OmapGetElement is bound to the ReadOp, if the ReadOp
// Release method is called the element may no longer be used.
type OmapGetElement struct {
	// inputs:
	startAfter   string
	filterPrefix string
	maxReturn    uint64

	// arguments:
	cStartAfter   *C.char
	cFilterPrefix *C.char

	// C returned data:
	iter C.rados_omap_iter_t
	more C.uchar
	rval C.int

	// internal state:

	// canIterate is only set after the operation is performed and is
	// intended to prevent premature fetching of data from the element
	canIterate bool
}

func newOmapGetElement(startAfter, filterPrefix string, maxReturn uint64) *OmapGetElement {
	oge := &OmapGetElement{
		startAfter:    startAfter,
		filterPrefix:  filterPrefix,
		maxReturn:     maxReturn,
		cStartAfter:   C.CString(startAfter),
		cFilterPrefix: C.CString(filterPrefix),
	}
	runtime.SetFinalizer(oge, freeElement)
	return oge
}

func (oge *OmapGetElement) free() {
	oge.reset()
	C.free(unsafe.Pointer(oge.cStartAfter))
	oge.cStartAfter = nil
	C.free(unsafe.Pointer(oge.cFilterPrefix))
	oge.cFilterPrefix = nil
}

func (oge *OmapGetElement) reset() {
	if oge.canIterate {
		C.rados_omap_get_end(oge.iter)
	}
	oge.canIterate = false
	oge.more = 0
	oge.rval = 0
}

func (oge *OmapGetElement) update() error {
	oge.canIterate = true
	return getError(oge.rval)
}

// Next returns the next key value pair or nil if iteration is exhausted.
func (oge *OmapGetElement) Next() (*OmapKeyValue, error) {
	if !oge.canIterate {
		return nil, ErrOperationIncomplete
	}
	var (
		cKey *C.char
		cVal *C.char
		cLen C.size_t
	)
	ret := C.rados_omap_get_next(oge.iter, &cKey, &cVal, &cLen)
	if ret != 0 {
		return nil, getError(ret)
	}
	if cKey == nil {
		return nil, nil
	}
	return &OmapKeyValue{
		Key:   C.GoString(cKey),
		Value: C.GoBytes(unsafe.Pointer(cVal), C.int(cLen)),
	}, nil
}

// More returns true if there are more matching keys available.
func (oge *OmapGetElement) More() bool {
	// tad bit hacky, but go can't automatically convert from
	// unsigned char to bool
	return oge.more != 0
}

// omapRmKeysElement is a write element used to track state, especially
// C memory, across the setup and use of a WriteOp.
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

// SetOmap appends the map `pairs` to the omap `oid`
func (ioctx *IOContext) SetOmap(oid string, pairs map[string][]byte) error {
	op := CreateWriteOp()
	defer op.Release()
	op.SetOmap(pairs)
	return op.operateCompat(ioctx, oid)
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
	op := CreateWriteOp()
	defer op.Release()
	op.CleanOmap()
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
func (ioctx *IOContext) ListOmapValues(
	oid string, startAfter string, filterPrefix string, maxReturn int64,
	listFn OmapListFunc) error {

	op := CreateReadOp()
	defer op.Release()
	ome := op.GetOmapValues(startAfter, filterPrefix, uint64(maxReturn))
	err := op.operateCompat(ioctx, oid)
	if err != nil {
		return err
	}

	for {
		kv, err := ome.Next()
		if err != nil {
			return err
		}
		if kv == nil {
			break
		}
		listFn(kv.Key, kv.Value)
	}
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
