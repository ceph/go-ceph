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

// Inode represents an inode number in the file system.
type Inode uint64

// DirEntry represents an entry within a directory.
type DirEntry struct {
	inode Inode
	name  string
}

// Name returns the directory entry's name.
func (d *DirEntry) Name() string {
	return d.name
}

// Inode returns the directory entry's inode number.
func (d *DirEntry) Inode() Inode {
	return d.inode
}

// toDirEntry converts a c struct dirent to our go wrapper.
func toDirEntry(de *C.struct_dirent) *DirEntry {
	return &DirEntry{
		inode: Inode(de.d_ino),
		name:  C.GoString(&de.d_name[0]),
	}
}

// ReadDir reads a single directory entry from the open Directory.
// A nil DirEntry pointer will be returned when the Directory stream has been
// exhausted.
//
// Implements:
//  int ceph_readdir_r(struct ceph_mount_info *cmount, struct ceph_dir_result *dirp, struct dirent *de);
func (dir *Directory) ReadDir() (*DirEntry, error) {
	var de C.struct_dirent
	ret := C.ceph_readdir_r(dir.mount.mount, dir.dir, &de)
	if ret < 0 {
		return nil, getError(ret)
	}
	if ret == 0 {
		return nil, nil // End-of-stream
	}
	return toDirEntry(&de), nil
}

// RewindDir sets the directory stream to the beginning of the directory.
//
// Implements:
//  void ceph_rewinddir(struct ceph_mount_info *cmount, struct ceph_dir_result *dirp);
func (dir *Directory) RewindDir() {
	C.ceph_rewinddir(dir.mount.mount, dir.dir)
}

// dirEntries provides a convenient wrapper around slices of DirEntry items.
// For example, use the Names() call to easily get only the names from a
// DirEntry slice.
type dirEntries []*DirEntry

// list returns all the contents of a directory as a dirEntries slice.
//
// list is implemented using ReadDir. If any of the calls to ReadDir returns
// an error List will return an error. However, all previous entries
// collected will still be returned. Callers of this function may want to check
// the entries return value even when an error is returned.
// List rewinds the handle every time it is called to get a full
// listing of directory contents.
func (dir *Directory) list() (dirEntries, error) {
	var (
		err     error
		entry   *DirEntry
		entries = make(dirEntries, 0)
	)
	dir.RewindDir()
	for {
		entry, err = dir.ReadDir()
		if err != nil || entry == nil {
			break
		}
		entries = append(entries, entry)
	}
	return entries, err
}

// names returns a slice of only the name fields from dir entries.
func (entries dirEntries) names() []string {
	names := make([]string, len(entries))
	for i, v := range entries {
		names[i] = v.Name()
	}
	return names
}
