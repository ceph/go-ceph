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

const (
	// SeekSet is used with Seek to set the absolute position in the file.
	SeekSet = int(C.SEEK_SET)
	// SeekCur is used with Seek to position the file relative to the current
	// position.
	SeekCur = int(C.SEEK_CUR)
	// SeekEnd is used with Seek to position the file relative to the end.
	SeekEnd = int(C.SEEK_END)
)

// File represents an open file descriptor in cephfs.
type File struct {
	mount *MountInfo
	fd    C.int
}

// Open a file at the given path. The flags are the same os flags as
// a local open call. Mode is the same mode bits as a local open call.
//
// Implements:
//  int ceph_open(struct ceph_mount_info *cmount, const char *path, int flags, mode_t mode);
func (mount *MountInfo) Open(path string, flags int, mode uint32) (*File, error) {
	if mount.mount == nil {
		return nil, ErrNotConnected
	}
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	ret := C.ceph_open(mount.mount, cPath, C.int(flags), C.mode_t(mode))
	if ret < 0 {
		return nil, getError(ret)
	}
	return &File{mount: mount, fd: ret}, nil
}

func (f *File) validate() error {
	if f.mount == nil {
		return ErrNotConnected
	}
	return nil
}

// Close the file.
//
// Implements:
//  int ceph_close(struct ceph_mount_info *cmount, int fd);
func (f *File) Close() error {
	if f.fd == -1 {
		// already closed
		return nil
	}
	if err := f.validate(); err != nil {
		return err
	}
	if err := getError(C.ceph_close(f.mount.mount, f.fd)); err != nil {
		return err
	}
	f.fd = -1
	return nil
}

// read directly wraps the ceph_read call. Because read is such a common
// operation we deviate from the ceph naming and expose Read and ReadAt
// wrappers for external callers of the library.
//
// Implements:
//  int ceph_read(struct ceph_mount_info *cmount, int fd, char *buf, int64_t size, int64_t offset);
func (f *File) read(buf []byte, offset int64) (int, error) {
	if err := f.validate(); err != nil {
		return 0, err
	}
	bufptr := (*C.char)(unsafe.Pointer(&buf[0]))
	ret := C.ceph_read(
		f.mount.mount, f.fd, bufptr, C.int64_t(len(buf)), C.int64_t(offset))
	if ret < 0 {
		return int(ret), getError(ret)
	}
	return int(ret), nil
}

// Read data from file. Up to len(buf) bytes will be read from the file.
// The number of bytes read will be returned.
func (f *File) Read(buf []byte) (int, error) {
	// to-consider: should we mimic Go's behavior of returning an
	// io.ErrShortWrite error if write length < buf size?
	return f.read(buf, -1)
}

// ReadAt will read data from the file starting at the given offset.
// Up to len(buf) bytes will be read from the file.
// The number of bytes read will be returned.
func (f *File) ReadAt(buf []byte, offset int64) (int, error) {
	return f.read(buf, offset)
}

// write directly wraps the ceph_write call. Because write is such a common
// operation we deviate from the ceph naming and expose Write and WriteAt
// wrappers for external callers of the library.
//
// Implements:
//  int ceph_write(struct ceph_mount_info *cmount, int fd, const char *buf,
//                 int64_t size, int64_t offset);
func (f *File) write(buf []byte, offset int64) (int, error) {
	if err := f.validate(); err != nil {
		return 0, err
	}
	bufptr := (*C.char)(unsafe.Pointer(&buf[0]))
	ret := C.ceph_write(
		f.mount.mount, f.fd, bufptr, C.int64_t(len(buf)), C.int64_t(offset))
	if ret < 0 {
		return 0, getError(ret)
	}
	return int(ret), nil
}

// Write data from buf to the file.
// The number of bytes written is returned.
func (f *File) Write(buf []byte) (int, error) {
	return f.write(buf, -1)
}

// WriteAt writes data from buf to the file at the specified offset.
// The number of bytes written is returned.
func (f *File) WriteAt(buf []byte, offset int64) (int, error) {
	return f.write(buf, offset)
}

// Seek will reposition the file stream based on the given offset.
//
// Implements:
//  int64_t ceph_lseek(struct ceph_mount_info *cmount, int fd, int64_t offset, int whence);
func (f *File) Seek(offset int64, whence int) (int64, error) {
	if err := f.validate(); err != nil {
		return 0, err
	}
	// validate the seek whence value in case the caller skews
	// from the seek values we technically support from C as documented.
	// TODO: need to support seek-(hole|data) in mimic and later.
	switch whence {
	case SeekSet, SeekCur, SeekEnd:
	default:
		return 0, errInvalid
	}

	ret := C.ceph_lseek(f.mount.mount, f.fd, C.int64_t(offset), C.int(whence))
	if ret < 0 {
		return 0, getError(C.int(ret))
	}
	return int64(ret), nil
}
