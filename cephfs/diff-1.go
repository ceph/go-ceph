//go:build ceph_preview

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
	"fmt"
	"unsafe"
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
	rawCephSnapDiffInfo := &C.struct_ceph_snapdiff_info{}
	ret := C.ceph_open_snapdiff(
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

	rawSnapDiffEntry := &C.struct_ceph_snapdiff_entry_t{}
	rawSnapDiffInfo := &C.struct_ceph_snapdiff_info{
		cmount:  info.CMount.mount,
		dir1:    info.Dir1.dir,
		dir_aux: info.DirAux.dir,
	}

	ret := C.ceph_readdir_snapdiff(
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

	rawCephSnapDiffInfo := &C.struct_ceph_snapdiff_info{
		cmount:  info.CMount.mount,
		dir1:    info.Dir1.dir,
		dir_aux: info.DirAux.dir,
	}
	ret := C.ceph_close_snapdiff(rawCephSnapDiffInfo)

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
