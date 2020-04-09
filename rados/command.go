package rados

// #cgo LDFLAGS: -lrados
// #include <stdlib.h>
// #include <rados/librados.h>
import "C"

import (
	"unsafe"
)

// MonCommand sends a command to one of the monitors
func (c *Conn) MonCommand(args []byte) ([]byte, string, error) {
	return c.monCommand(args, nil)
}

// MonCommandWithInputBuffer sends a command to one of the monitors, with an input buffer
func (c *Conn) MonCommandWithInputBuffer(args, inputBuffer []byte) ([]byte, string, error) {
	return c.monCommand(args, inputBuffer)
}

func (c *Conn) monCommand(args, inputBuffer []byte) ([]byte, string, error) {
	argv := C.CString(string(args))
	defer C.free(unsafe.Pointer(argv))

	var (
		info               string
		buffer             []byte
		outs, outbuf       *C.char
		outslen, outbuflen C.size_t
	)
	inbuf := C.CString(string(inputBuffer))
	inbufLen := len(inputBuffer)
	defer C.free(unsafe.Pointer(inbuf))

	ret := C.rados_mon_command(c.cluster,
		&argv, 1,
		inbuf,              // bulk input (e.g. crush map)
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
		err := getError(ret)
		return nil, info, err
	}

	return buffer, info, nil
}

// PGCommand sends a command to one of the PGs
//
// Implements:
//  int rados_pg_command(rados_t cluster, const char *pgstr,
//                       const char **cmd, size_t cmdlen,
//                       const char *inbuf, size_t inbuflen,
//                       char **outbuf, size_t *outbuflen,
//                       char **outs, size_t *outslen);
func (c *Conn) PGCommand(pgid []byte, args [][]byte) ([]byte, string, error) {
	return c.pgCommand(pgid, args, nil)
}

// PGCommandWithInputBuffer sends a command to one of the PGs, with an input buffer
//
// Implements:
//  int rados_pg_command(rados_t cluster, const char *pgstr,
//                       const char **cmd, size_t cmdlen,
//                       const char *inbuf, size_t inbuflen,
//                       char **outbuf, size_t *outbuflen,
//                       char **outs, size_t *outslen);
func (c *Conn) PGCommandWithInputBuffer(pgid []byte, args [][]byte, inputBuffer []byte) ([]byte, string, error) {
	return c.pgCommand(pgid, args, inputBuffer)
}

func (c *Conn) pgCommand(pgid []byte, args [][]byte, inputBuffer []byte) ([]byte, string, error) {
	name := C.CString(string(pgid))
	defer C.free(unsafe.Pointer(name))

	argc := len(args)
	argv := make([]*C.char, argc)

	for i, arg := range args {
		argv[i] = C.CString(string(arg))
		defer C.free(unsafe.Pointer(argv[i]))
	}

	var (
		info               string
		buffer             []byte
		outs, outbuf       *C.char
		outslen, outbuflen C.size_t
	)
	inbuf := C.CString(string(inputBuffer))
	inbufLen := len(inputBuffer)
	defer C.free(unsafe.Pointer(inbuf))

	ret := C.rados_pg_command(c.cluster,
		name,
		&argv[0],
		C.size_t(argc),
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
		err := getError(ret)
		return nil, info, err
	}

	return buffer, info, nil
}

// MgrCommand sends a command to a ceph-mgr.
func (c *Conn) MgrCommand(args [][]byte) ([]byte, string, error) {
	return c.mgrCommand(args, nil)
}

// MgrCommandWithInputBuffer sends a command, with an input buffer, to a ceph-mgr.
func (c *Conn) MgrCommandWithInputBuffer(args [][]byte, inputBuffer []byte) ([]byte, string, error) {
	return c.mgrCommand(args, inputBuffer)
}

// Implements:
//  int rados_mgr_command(rados_t cluster, const char **cmd,
//                         size_t cmdlen, const char *inbuf,
//                         size_t inbuflen, char **outbuf,
//                         size_t *outbuflen, char **outs,
//                          size_t *outslen);
func (c *Conn) mgrCommand(args [][]byte, inputBuffer []byte) ([]byte, string, error) {
	argc := len(args)
	argv := make([]*C.char, argc)

	for i, arg := range args {
		argv[i] = C.CString(string(arg))
		defer C.free(unsafe.Pointer(argv[i]))
	}

	var (
		info               string
		buffer             []byte
		outs, outbuf       *C.char
		outslen, outbuflen C.size_t
	)
	inbuf := C.CString(string(inputBuffer))
	inbufLen := len(inputBuffer)
	defer C.free(unsafe.Pointer(inbuf))

	ret := C.rados_mgr_command(
		c.cluster,
		&argv[0],
		C.size_t(argc),
		inbuf,
		C.size_t(inbufLen),
		&outbuf,
		&outbuflen,
		&outs,
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
		err := getError(ret)
		return nil, info, err
	}

	return buffer, info, nil
}
