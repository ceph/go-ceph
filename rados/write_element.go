package rados

// #include <stdint.h>
import "C"

import (
	"unsafe"
)

type writeElement struct {
	// inputs:
	b []byte

	// arguments:
	cBuffer   *C.char
	cDataLen  C.size_t
	cWriteLen C.size_t
	cOffset   C.uint64_t
}

func newWriteElement(b []byte, writeLen, offset uint64) *writeElement {
	return &writeElement{
		b:         b,
		cBuffer:   (*C.char)(unsafe.Pointer(&b[0])),
		cDataLen:  C.size_t(len(b)),
		cWriteLen: C.size_t(writeLen),
		cOffset:   C.uint64_t(offset),
	}
}

func (*writeElement) reset()        {}
func (*writeElement) update() error { return nil }
func (*writeElement) free()         {}
