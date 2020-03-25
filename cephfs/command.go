package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <cephfs/libcephfs.h>
*/
import "C"

import (
	"unsafe"
)

// MdsCommand sends commands to the specified MDS.
func (mount *MountInfo) MdsCommand(mdsSpec string, args [][]byte) ([]byte, string, error) {
	return mount.mdsCommand(mdsSpec, args, nil)
}

// MdsCommandWithInputBuffer sends commands to the specified MDS, with an input
// buffer.
func (mount *MountInfo) MdsCommandWithInputBuffer(mdsSpec string, args [][]byte, inputBuffer []byte) ([]byte, string, error) {
	return mount.mdsCommand(mdsSpec, args, inputBuffer)
}

// mdsCommand supports sending formatted commands to MDS.
//
// Implements:
//  int ceph_mds_command(struct ceph_mount_info *cmount,
//      const char *mds_spec,
//      const char **cmd,
//      size_t cmdlen,
//      const char *inbuf, size_t inbuflen,
//      char **outbuf, size_t *outbuflen,
//      char **outs, size_t *outslen);
func (mount *MountInfo) mdsCommand(mdsSpec string, args [][]byte, inputBuffer []byte) (buffer []byte, info string, err error) {
	spec := C.CString(mdsSpec)
	defer C.free(unsafe.Pointer(spec))

	argc := len(args)
	argv := make([]*C.char, argc)

	for i, arg := range args {
		argv[i] = C.CString(string(arg))
	}
	// free all array elements in a single defer
	defer func() {
		for i := range argv {
			C.free(unsafe.Pointer(argv[i]))
		}
	}()

	var (
		outs, outbuf       *C.char
		outslen, outbuflen C.size_t
	)
	inbuf := C.CString(string(inputBuffer))
	inbufLen := len(inputBuffer)
	defer C.free(unsafe.Pointer(inbuf))

	ret := C.ceph_mds_command(
		mount.mount,        // cephfs mount ref
		spec,               // mds spec
		&argv[0],           // cmd array
		C.size_t(argc),     // cmd array length
		inbuf,              // bulk input
		C.size_t(inbufLen), // length inbuf
		&outbuf,            // buffer
		&outbuflen,         // buffer length
		&outs,              // status string
		&outslen)

	if outslen > 0 {
		info = C.GoStringN(outs, C.int(outslen))
		C.free(unsafe.Pointer(outs))
	}
	if outbuflen > 0 {
		buffer = C.GoBytes(unsafe.Pointer(outbuf), C.int(outbuflen))
		C.free(unsafe.Pointer(outbuf))
	}
	if ret != 0 {
		return nil, info, getError(ret)
	}

	return buffer, info, nil
}
