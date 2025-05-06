//go:build ceph_preview

package cephfs

/*
#cgo LDFLAGS: -lcephfs
#cgo CPPFLAGS: -D_FILE_OFFSET_BITS=64
#include <stdlib.h>
#include <dirent.h>
#include <cephfs/libcephfs.h>

// open_snapdiff_fn matches the open_snapdiff function signature.
typedef int(*open_snapdiff_fn)(struct ceph_mount_info* cmount,
                                  const char* root_path,
                                  const char* rel_path,
                                  const char* snap1,
                                  const char* snap2,
                                  struct ceph_snapdiff_info* out);

// open_snapdiff_dlsym take *fn as open_snapdiff_fn and calls the dynamically loaded
// open_snapdiff function passed as 1st argument.
static inline int open_snapdiff_dlsym(void *fn,
                                  struct ceph_mount_info* cmount,
                                  const char* root_path,
                                  const char* rel_path,
                                  const char* snap1,
                                  const char* snap2,
                                  struct ceph_snapdiff_info* out) {
	// cast function pointer fn to open_snapdiff and call the function
	return ((open_snapdiff_fn) fn)(cmount, root_path, rel_path, snap1, snap2, out);
}

// readdir_snapdiff_fn matches the readdir_snapdiff function signature.
typedef int(*readdir_snapdiff_fn)(struct ceph_snapdiff_info* snapdiff,
                                  struct ceph_snapdiff_entry_t* out);

// readdir_snapdiff_dlsym take *fn as readdir_snapdiff_fn and calls the dynamically loaded
// readdir_snapdiff function passed as 1st argument.
static inline int readdir_snapdiff_dlsym(void *fn,
                                  struct ceph_snapdiff_info* snapdiff,
                                  struct ceph_snapdiff_entry_t* out) {
	// cast function pointer fn to readdir_snapdiff and call the function
	return ((readdir_snapdiff_fn) fn)(snapdiff, out);
}

// close_snapdiff_fn matches the close_snapdiff function signature.
typedef int(*close_snapdiff_fn)(struct ceph_snapdiff_info* snapdiff);

// close_snapdiff_dlsym take *fn as close_snapdiff_fn and calls the dynamically loaded
// close_snapdiff function passed as 1st argument.
static inline int close_snapdiff_dlsym(void *fn,
                                  struct ceph_snapdiff_info* snapdiff) {
	// cast function pointer fn to close_snapdiff and call the function
	return ((close_snapdiff_fn) fn)(snapdiff);
}
*/
import "C"

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ceph/go-ceph/internal/dlsym"
)

var (
	cephOpenSnapDiffOnce    sync.Once
	cephReaddirSnapDiffOnce sync.Once
	cephCloseSnapDiffOnce   sync.Once
	cephOpenSnapDiff        unsafe.Pointer
	cephReaddirSnapDiff     unsafe.Pointer
	cephCloseSnapDiff       unsafe.Pointer
	cephOpenSnapDiffErr     error
	cephReaddirSnapDiffErr  error
	cephCloseSnapDiffErr    error
)

// struct ceph_snapdiff_info
//
//	{
//	  struct ceph_mount_info* cmount;
//	  struct ceph_dir_result* dir1;    // primary dir entry to build snapdiff for.
//	  struct ceph_dir_result* dir_aux; // aux dir entry to identify the second snapshot.
//	                                   // Can point to the parent dir entry if entry-in-question
//	                                   // doesn't exist in the second snapshot
//	};
type SnapDiffInfo struct {
	CMount *MountInfo
	Dir1   *Directory
	DirAux *Directory
}

// SnapDiffEntry is a single entry in the snapshot diff.
//
//	struct ceph_snapdiff_entry_t {
//			struct dirent dir_entry;
//			uint64_t snapid; //should be snapid_t but prefer not to exposure it
//	 };
type SnapDiffEntry struct {
	DirEntry *DirEntry
	SnapID   uint64
}

// SnapDiffConfig is used to define the parameters of a open_snapdiff call.

type SnapDiffConfig struct {
	CMount   *MountInfo
	RootPath string
	RelPath  string
	Snap1    string
	Snap2    string
}

// CephOpenSnapDiff opens a snapshot diff between two snapshots of a file
// and returns a SnapDiffInfo struct containing the diff information.
//
// Implements:
//
//	int ceph_open_snapdiff(struct ceph_mount_info* cmount,
//	                       const char* root_path,
//	                       const char* rel_path,
//	                       const char* snap1,
//	                       const char* snap2,
//	                       struct ceph_snapdiff_info* out);
func CephOpenSnapDiff(config SnapDiffConfig) (*SnapDiffInfo, error) {
	// if config.CMount == nil || config.RootPath == "" || config.RelPath == "" ||
	// 	config.Snap1 == "" || config.Snap2 == "" {
	// 	return nil, fmt.Errorf("invalid arguments")
	// }

	cephOpenSnapDiffOnce.Do(func() {
		cephOpenSnapDiff, cephOpenSnapDiffErr = dlsym.LookupSymbol("ceph_open_snapdiff")
	})

	if cephOpenSnapDiffErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotImplemented, cephOpenSnapDiffErr)
	}

	rawCephSnapDiffInfo := &C.struct_ceph_snapdiff_info{}

	ret := C.open_snapdiff_dlsym(
		cephOpenSnapDiff,
		config.CMount.mount,
		C.CString(config.RootPath),
		C.CString(config.RelPath),
		C.CString(config.Snap1),
		C.CString(config.Snap2),
		rawCephSnapDiffInfo)

	if ret != 0 {
		return nil, getError(ret)
	}

	mountInfo := &MountInfo{
		mount: rawCephSnapDiffInfo.cmount,
	}
	cephSnapDiffInfo := &SnapDiffInfo{
		CMount: mountInfo,
		Dir1: &Directory{
			mount: mountInfo,
			dir:   rawCephSnapDiffInfo.dir1,
		},
		DirAux: &Directory{
			mount: mountInfo,
			dir:   rawCephSnapDiffInfo.dir_aux,
		},
	}

	return cephSnapDiffInfo, nil
}

// CephReaddirSnapDiff reads the next entry in the snapshot diff.
// The entry is returned in the out parameter.
//
// Implements:
//
//	int ceph_readdir_snapdiff(struct ceph_snapdiff_info* snapdiff,
//	                           struct ceph_snapdiff_entry_t* out);
func CephReaddirSnapDiff(info *SnapDiffInfo) (*SnapDiffEntry, error) {
	if info == nil {
		return nil, nil
	}

	cephReaddirSnapDiffOnce.Do(func() {
		cephReaddirSnapDiff, cephReaddirSnapDiffErr = dlsym.LookupSymbol("ceph_readdir_snapdiff")
	})
	if cephReaddirSnapDiffErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotImplemented, cephReaddirSnapDiffErr)
	}

	rawSnapDiffEntry := &C.struct_ceph_snapdiff_entry_t{}
	rawSnapDiffInfo := &C.struct_ceph_snapdiff_info{
		cmount:  info.CMount.mount,
		dir1:    info.Dir1.dir,
		dir_aux: info.DirAux.dir,
	}

	ret := C.readdir_snapdiff_dlsym(
		cephReaddirSnapDiff,
		rawSnapDiffInfo,
		rawSnapDiffEntry)
	if ret != 0 {
		return nil, getError(ret)
	}
	snapDiffEntry := &SnapDiffEntry{
		DirEntry: toDirEntry(&rawSnapDiffEntry.dir_entry),
		SnapID:   uint64(rawSnapDiffEntry.snapid),
	}
	return snapDiffEntry, nil
}

// CephCloseSnapDiff closes the snapshot diff handle.
//
// Implements:
//
//	int ceph_close_snapdiff(struct ceph_snapdiff_info* snapdiff);
func CephCloseSnapDiff(info *SnapDiffInfo) error {
	if info == nil {
		return nil
	}

	cephCloseSnapDiffOnce.Do(func() {
		cephCloseSnapDiff, cephCloseSnapDiffErr = dlsym.LookupSymbol("ceph_close_snapdiff")
	})
	if cephCloseSnapDiffErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, cephCloseSnapDiffErr)
	}

	rawCephSnapDiffInfo := &C.struct_ceph_snapdiff_info{
		cmount:  info.CMount.mount,
		dir1:    info.Dir1.dir,
		dir_aux: info.DirAux.dir,
	}
	ret := C.close_snapdiff_dlsym(
		cephCloseSnapDiff,
		rawCephSnapDiffInfo)

	if ret != 0 {
		return getError(ret)
	}

	return nil
}

func GetSnapshotID(cMount *MountInfo, path string) (uint64, error) {
	if cMount == nil || path == "" {
		return 0, fmt.Errorf("invalid arguments")
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	snapInfo := &C.struct_snap_info{}

	ret := C.ceph_get_snap_info(cMount.mount, cPath, snapInfo)
	if ret != 0 {
		return 0, getError(ret)
	}
	snapID := uint64(snapInfo.id)

	C.ceph_free_snap_info_buffer(snapInfo)

	return snapID, nil
}
