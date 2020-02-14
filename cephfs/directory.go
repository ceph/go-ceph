package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <dirent.h>
#include <cephfs/libcephfs.h>
*/
import "C"

import (
	"unsafe"
)

// Directory represents an open directory handle.
type Directory struct {
	mount *MountInfo
	dir   *C.struct_ceph_dir_result
}

// OpenDir returns a new Directory handle open for I/O.
//
// Implements:
//  int ceph_opendir(struct ceph_mount_info *cmount, const char *name, struct ceph_dir_result **dirpp);
func (mount *MountInfo) OpenDir(path string) (*Directory, error) {
	var dir *C.struct_ceph_dir_result

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	ret := C.ceph_opendir(mount.mount, cPath, &dir)
	if ret != 0 {
		return nil, getError(ret)
	}

	return &Directory{
		mount: mount,
		dir:   dir,
	}, nil
}

// Close the open directory handle.
//
// Implements:
//  int ceph_closedir(struct ceph_mount_info *cmount, struct ceph_dir_result *dirp);
func (dir *Directory) Close() error {
	return getError(C.ceph_closedir(dir.mount.mount, dir.dir))
}
