//go:build ceph_preview

package rbd

/*
#cgo LDFLAGS: -lrbd
#include <errno.h>
#include <stdlib.h>
#include <rados/librados.h>
#include <rbd/librbd.h>

typedef enum {
  RBD_MIRROR_GROUP_DISABLING = 0,
  RBD_MIRROR_GROUP_ENABLING = 1,
  RBD_MIRROR_GROUP_ENABLED = 2,
  RBD_MIRROR_GROUP_DISABLED = 3
} rbd_mirror_group_state_t;

typedef enum {
  MIRROR_GROUP_STATUS_STATE_UNKNOWN         = 0,
  MIRROR_GROUP_STATUS_STATE_ERROR           = 1,
  MIRROR_GROUP_STATUS_STATE_STARTING_REPLAY = 2,
  MIRROR_GROUP_STATUS_STATE_REPLAYING       = 3,
  MIRROR_GROUP_STATUS_STATE_STOPPING_REPLAY = 4,
  MIRROR_GROUP_STATUS_STATE_STOPPED         = 5,
} rbd_mirror_group_status_state_t;

typedef struct {
  char *global_id;
  rbd_mirror_image_mode_t mirror_image_mode;
  rbd_mirror_group_state_t state;
  bool primary;
} rbd_mirror_group_info_t;

typedef struct {
  char *mirror_uuid;
  rbd_mirror_group_status_state_t state;
  char *description;
  uint32_t mirror_image_count;
  int64_t *mirror_image_pool_ids;
  char **mirror_image_global_ids;
  rbd_mirror_image_site_status_t *mirror_images;
  time_t last_update;
  bool up;
} rbd_mirror_group_site_status_t;

typedef struct {
  char *name;
  rbd_mirror_group_info_t info;
  uint32_t site_statuses_count;
  rbd_mirror_group_site_status_t *site_statuses;
} rbd_mirror_group_global_status_t;

// rbd_mirror_group_enable_fn matches rbd_mirror_group_enable function.
typedef int(*rbd_mirror_group_enable_fn)(rados_ioctx_t p, const char *name,
                                         rbd_mirror_image_mode_t mirror_image_mode, uint32_t flags);

// rbd_mirror_group_enable_dlsym take *fn as rbd_mirror_group_enable_fn and calls the dynamically loaded
// rbd_mirror_group_enable function passed as 1st argument.
static inline int rbd_mirror_group_enable_dlsym(void *fn, rados_ioctx_t p, const char *name,
                                               rbd_mirror_image_mode_t mirror_image_mode, uint32_t flags) {
  // cast function pointer fn to rbd_mirror_group_enable and call the function
  return ((rbd_mirror_group_enable_fn) fn)(p, name, mirror_image_mode, flags);
}

// rbd_mirror_group_disable_fn matches rbd_mirror_group_disable function.
typedef int(*rbd_mirror_group_disable_fn)(rados_ioctx_t p, const char *name, bool force);

// rbd_mirror_group_disable_dlsym take *fn as rbd_mirror_group_disable_fn and calls the dynamically loaded
// rbd_mirror_group_disable function passed as 1st argument.
static inline int rbd_mirror_group_disable_dlsym(void *fn, rados_ioctx_t p, const char *name, bool force) {
  // cast function pointer fn to rbd_mirror_group_disable and call the function
  return ((rbd_mirror_group_disable_fn) fn)(p, name, force);
}

// rbd_mirror_group_promote_fn matches rbd_mirror_group_promote function.
typedef int(*rbd_mirror_group_promote_fn)(rados_ioctx_t p, const char *name, uint32_t flags, bool force);

// rbd_mirror_group_promote_dlsym take *fn as rbd_mirror_group_promote_fn and calls the dynamically loaded
//rbd_mirror_group_promote function passed as 1st argument.
static inline int rbd_mirror_group_promote_dlsym(void *fn, rados_ioctx_t p, const char *name,
                                                 uint32_t flags, bool force){
  // cast function pointer fn to rbd_mirror_group_promote and call the function
  return ((rbd_mirror_group_promote_fn) fn)(p, name, flags, force);
}

// rbd_mirror_group_demote_fn matches rbd_mirror_group_demote function.
typedef int(*rbd_mirror_group_demote_fn)(rados_ioctx_t p, const char *name, uint32_t flags);

// rbd_mirror_group_demote_dlsym take *fn as rbd_mirror_group_demote_fn and calls the dynamically loaded
// rbd_mirror_group_demote function passed as 1st argument.
static inline int rbd_mirror_group_demote_dlsym(void *fn, rados_ioctx_t p, const char *name, uint32_t flags){
  // cast function pointer fn to rbd_mirror_group_demote and call the function
  return ((rbd_mirror_group_demote_fn) fn)(p, name, flags);
}

// rbd_mirror_group_resync_fn matches rbd_mirror_group_resync function.
typedef int(*rbd_mirror_group_resync_fn)(rados_ioctx_t p, const char *name);

// rbd_mirror_group_resync_dlsym take *fn as rbd_mirror_group_resync_fn and calls the dynamically loaded
// rbd_mirror_group_resync function passed as 1st argument.
static inline int rbd_mirror_group_resync_dlsym(void *fn, rados_ioctx_t p, const char *name){
  // cast function pointer fn to rbd_mirror_group_resync and call the function
  return ((rbd_mirror_group_resync_fn) fn)(p, name);
}

// rbd_mirror_group_get_info_fn matches rbd_mirror_group_get_info function.
typedef int(*rbd_mirror_group_get_info_fn)(rados_ioctx_t p, const char *name,
                                           rbd_mirror_group_global_status_t *mirror_group_status,
                                           size_t status_size);

// rbd_mirror_group_get_info_dlsym take *fn as rbd_mirror_group_get_info_fn and calls the dynamically loaded
// rbd_mirror_group_get_info function passed as 1st argument.
static inline int rbd_mirror_group_get_info_dlsym(void *fn, rados_ioctx_t p, const char *name,
                                                  rbd_mirror_group_global_status_t *mirror_group_status,
                                                  size_t status_size){
  // cast function pointer fn to rbd_mirror_group_get_info and call the function
  return ((rbd_mirror_group_get_info_fn) fn)(p, name, mirror_group_status, status_size);
}

// rbd_mirror_group_get_global_status_fn matches rbd_mirror_group_get_global_status function.
typedef int(*rbd_mirror_group_get_global_status_fn)(rados_ioctx_t p, const char *name,
                                                    rbd_mirror_group_info_t *mirror_group_info,
                                                    size_t info_size);

// rbd_mirror_group_get_global_status_dlsym take *fn as rbd_mirror_group_get_global_status_fn and calls the dynamically loaded
// rbd_mirror_group_get_global_status function passed as 1st argument.
static inline int rbd_mirror_group_get_global_status_dlsym(void *fn, rados_ioctx_t p, const char *name,
                                                           rbd_mirror_group_info_t *mirror_group_info,
                                                            size_t info_size){
  // cast function pointer fn to rbd_mirror_group_get_global_status and call the function
  return ((rbd_mirror_group_get_global_status_fn) fn)(p, name, mirror_group_info, info_size);
}
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/ceph/go-ceph/internal/cutil"
	"github.com/ceph/go-ceph/internal/dlsym"
	"github.com/ceph/go-ceph/rados"
)

// MirrorGroupEnable will enable mirroring for a group using the specified mode.
//
// Implements:
//
//	int rbd_mirror_group_enable(rados_ioctx_t p, const char *name,
//	  							rbd_mirror_image_mode_t mirror_image_mode,
//									uint32_t flags);
func MirrorGroupEnable(groupIoctx *rados.IOContext, groupName string, mode ImageMirrorMode) error {
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))

	rbdMirrorGroupEnable, rbdMirrorGroupEnableErr := dlsym.LookupSymbol("rbd_mirror_group_enable")
	if rbdMirrorGroupEnableErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, rbdMirrorGroupEnableErr)
	}

	ret := C.rbd_mirror_group_enable_dlsym(
		rbdMirrorGroupEnable,
		cephIoctx(groupIoctx),
		cGroupName,
		C.rbd_mirror_image_mode_t(mode),
		(C.uint32_t)(2),
	)

	return getError(ret)
}

// MirrorGroupDisable will disabling mirroring for a group
//
// Implements:
//
//	int rbd_mirror_group_disable(rados_ioctx_t p, const char *name,
//	  							bool force)
func MirrorGroupDisable(groupIoctx *rados.IOContext, groupName string, force bool) error {
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))

	rbdMirrorGroupDisable, rbdMirrorGroupDisableErr := dlsym.LookupSymbol("rbd_mirror_group_disable")
	if rbdMirrorGroupDisableErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, rbdMirrorGroupDisableErr)
	}

	ret := C.rbd_mirror_group_disable_dlsym(
		rbdMirrorGroupDisable,
		cephIoctx(groupIoctx),
		cGroupName,
		C.bool(force))

	return getError(ret)
}

// MirrorGroupPromote will promote the mirrored group to primary status
//
// Implements:
//
//	int rbd_mirror_group_promote(rados_ioctx_t p, const char *name,
//	  							uint32_t flags, bool force)
func MirrorGroupPromote(groupIoctx *rados.IOContext, groupName string, force bool) error {
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))

	rbdMirrorGroupPromote, rbdMirrorGroupPromoteErr := dlsym.LookupSymbol("rbd_mirror_group_promote")
	if rbdMirrorGroupPromoteErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, rbdMirrorGroupPromoteErr)
	}

	ret := C.rbd_mirror_group_promote_dlsym(
		rbdMirrorGroupPromote,
		cephIoctx(groupIoctx),
		cGroupName,
		(C.uint32_t)(0),
		C.bool(force))

	return getError(ret)
}

// MirrorGroupDemote will demote the mirrored group to primary status
//
// Implements:
//
//	int rbd_mirror_group_demote(rados_ioctx_t p, const char *name,
//	  							uint32_t flags)
func MirrorGroupDemote(groupIoctx *rados.IOContext, groupName string) error {
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))

	rbdMirrorGroupDemote, rbdMirrorGroupDemoteErr := dlsym.LookupSymbol("rbd_mirror_group_demote")
	if rbdMirrorGroupDemoteErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, rbdMirrorGroupDemoteErr)
	}

	ret := C.rbd_mirror_group_demote_dlsym(
		rbdMirrorGroupDemote,
		cephIoctx(groupIoctx),
		cGroupName,
		(C.uint32_t)(0))
	return getError(ret)
}

// MirrorGroupResync is used to manually resolve split-brain status by triggering
// resynchronization
//
// Implements:
//
//	int rbd_mirror_group_resync(rados_ioctx_t p, const char *name)
func MirrorGroupResync(groupIoctx *rados.IOContext, groupName string) error {
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))

	rbdMirrorGroupResync, rbdMirrorGroupResyncErr := dlsym.LookupSymbol("rbd_mirror_group_resync")
	if rbdMirrorGroupResyncErr != nil {
		return fmt.Errorf("%w: %w", ErrNotImplemented, rbdMirrorGroupResyncErr)
	}

	ret := C.rbd_mirror_group_resync_dlsym(
		rbdMirrorGroupResync,
		cephIoctx(groupIoctx),
		cGroupName)
	return getError(ret)
}

// MirrorGroupState represents the current state of the mirrored group
type MirrorGroupState C.rbd_mirror_group_state_t

// String representation of MirrorGroupState.
func (mgs MirrorGroupState) String() string {
	switch mgs {
	case MirrorGroupEnabled:
		return "enabled"
	case MirrorGroupDisabled:
		return "disabled"
	case MirrorGroupEnabling:
		return "enabling"
	case MirrorGrpupDisabling:
		return "disabled"
	default:
		return "<unknown>"
	}
}

const (
	// MirrorGrpupDisabling is the representation of
	// RBD_MIRROR_GROUP_DISABLING from librbd.
	MirrorGrpupDisabling = MirrorGroupState(C.RBD_MIRROR_GROUP_DISABLING)
	// MirrorGroupEnabling is the representation of
	// RBD_MIRROR_GROUP_ENABLING from librbd
	MirrorGroupEnabling = MirrorGroupState(C.RBD_MIRROR_GROUP_ENABLING)
	// MirrorGroupEnabled is the representation of
	// RBD_MIRROR_IMAGE_ENABLED from librbd.
	MirrorGroupEnabled = MirrorGroupState(C.RBD_MIRROR_GROUP_ENABLED)
	// MirrorGroupDisabled is the representation of
	// RBD_MIRROR_GROUP_DISABLED from librbd.
	MirrorGroupDisabled = MirrorGroupState(C.RBD_MIRROR_GROUP_DISABLED)
)

// MirrorGroupInfo represents the mirroring status information of group.
type MirrorGroupInfo struct {
	GlobalID        string
	State           MirrorGroupState
	MirrorImageMode ImageMirrorMode
	Primary         bool
}

// GetMirrorGroupInfo returns the mirroring status information of the mirrored group
//
// Implements:
//
//	int rbd_mirror_group_get_info(rados_ioctx_t p, const char *name,
//								  rbd_mirror_group_info_t *mirror_group_info,
//								  size_t info_size)
func GetMirrorGroupInfo(groupIoctx *rados.IOContext, groupName string) (*MirrorGroupInfo, error) {
	var cgInfo C.rbd_mirror_group_info_t
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))

	rbdMirrorGroupGetInfo, rbdMirrorGroupGetInfoErr := dlsym.LookupSymbol("rbd_mirror_group_get_info")
	if rbdMirrorGroupGetInfoErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotImplemented, rbdMirrorGroupGetInfoErr)
	}

	ret := C.rbd_mirror_group_get_info_dlsym(
		rbdMirrorGroupGetInfo,
		cephIoctx(groupIoctx),
		cGroupName,
		&cgInfo,
		C.sizeof_rbd_mirror_group_info_t)

	if ret < 0 {
		return nil, getError(ret)
	}

	info := convertMirrorGroupInfo(&cgInfo)

	return &info, nil

}

func convertMirrorGroupInfo(cgInfo *C.rbd_mirror_group_info_t) MirrorGroupInfo {
	return MirrorGroupInfo{
		GlobalID:        C.GoString(cgInfo.global_id),
		MirrorImageMode: ImageMirrorMode(cgInfo.mirror_image_mode),
		State:           MirrorGroupState(cgInfo.state),
		Primary:         bool(cgInfo.primary),
	}
}

// MirrorGroupStatusState is used to indicate the state of a mirrored group
// within the site status info.
type MirrorGroupStatusState int64

const (
	// MirrorGroupStatusStateUnknown is equivalent to MIRROR_GROUP_STATUS_STATE_UNKNOWN
	MirrorGroupStatusStateUnknown = MirrorGroupStatusState(C.MIRROR_GROUP_STATUS_STATE_UNKNOWN)
	// MirrorGroupStatusStateError is equivalent to MIRROR_GROUP_STATUS_STATE_ERROR
	MirrorGroupStatusStateError = MirrorGroupStatusState(C.MIRROR_GROUP_STATUS_STATE_ERROR)
	// MirrorGroupStatusStateStartingReplay is equivalent to MIRROR_GROUP_STATUS_STATE_STARTING_REPLAY
	MirrorGroupStatusStateStartingReplay = MirrorGroupStatusState(C.MIRROR_GROUP_STATUS_STATE_STARTING_REPLAY)
	// MirrorGroupStatusStateReplaying is equivalent to MIRROR_GROUP_STATUS_STATE_REPLAYING
	MirrorGroupStatusStateReplaying = MirrorGroupStatusState(C.MIRROR_GROUP_STATUS_STATE_REPLAYING)
	// MirrorGroupStatusStateStoppingReplay is equivalent to MIRROR_GROUP_STATUS_STATE_STOPPING_REPLAY
	MirrorGroupStatusStateStoppingReplay = MirrorGroupStatusState(C.MIRROR_GROUP_STATUS_STATE_STOPPING_REPLAY)
	// MirrorGroupStatusStateStopped is equivalent to MIRROR_IMAGE_GROUP_STATUS_STATE_STOPPED
	MirrorGroupStatusStateStopped = MirrorGroupStatusState(C.MIRROR_GROUP_STATUS_STATE_STOPPED)
)

// SiteMirrorGroupStatus contains information pertaining to the status of
// a mirrored group within a site.
type SiteMirrorGroupStatus struct {
	MirrorUUID           string
	State                MirrorGroupStatusState
	MirrorImageCount     int
	MirrorImagePoolIDs   int64
	MirrorImageGlobalIDs string
	MirrorImages         []SiteMirrorImageStatus
	Description          string
	LastUpdate           int64
	Up                   bool
}

// GlobalMirrorGroupStatus contains information pertaining to the global
// status of a mirrored group. It contains general information as well
// as per-site information stored in the SiteStatuses slice.
type GlobalMirrorGroupStatus struct {
	Name              string
	Info              MirrorGroupInfo
	SiteStatusesCount int
	SiteStatuses      []SiteMirrorGroupStatus
}

type groupSiteArray [cutil.MaxIdx]C.rbd_mirror_group_site_status_t

// GetGlobalMirrorGroupStatus returns status information pertaining to the state
// of a groups's mirroring.
//
// Implements:
//
//	int rbd_mirror_group_get_global_status(
//		rados_ioctx_t p,
//		const char *group_name
//		rbd_mirror_group_global_status_t *mirror_group_status,
//		size_t status_size);
func GetGlobalMirrorGroupStatus(ioctx *rados.IOContext, groupName string) (GlobalMirrorGroupStatus, error) {
	s := C.rbd_mirror_group_global_status_t{}
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))

	rbdMirrorGroupGetGlobalStatus, rbdMirrorGroupGetGlobalStatusErr := dlsym.LookupSymbol("rbd_mirror_group_get_global_status")
	if rbdMirrorGroupGetGlobalStatusErr != nil {
		return nil, fmt.Errorf("%w: %w", ErrNotImplemented, rbdMirrorGroupGetGlobalStatusErr)
	}

	ret := C.rbd_mirror_group_get_global_status_dlsym(
		rbdMirrorGroupGetGlobalStatus,
		cephIoctx(ioctx),
		(*C.char)(cGroupName),
		&s,
		C.sizeof_rbd_mirror_group_global_status_t)
	if err := getError(ret); err != nil {
		return GlobalMirrorGroupStatus{}, err
	}

	status := newGlobalMirrorGroupStatus(&s)
	return status, nil
}

func newGlobalMirrorGroupStatus(
	s *C.rbd_mirror_group_global_status_t) GlobalMirrorGroupStatus {

	status := GlobalMirrorGroupStatus{
		Name:              C.GoString(s.name),
		Info:              convertMirrorGroupInfo(&s.info),
		SiteStatusesCount: int(s.site_statuses_count),
		SiteStatuses:      make([]SiteMirrorGroupStatus, s.site_statuses_count),
	}
	gsscs := (*groupSiteArray)(unsafe.Pointer(s.site_statuses))[:s.site_statuses_count:s.site_statuses_count]
	for i := C.uint32_t(0); i < s.site_statuses_count; i++ {
		gss := gsscs[i]
		status.SiteStatuses[i] = SiteMirrorGroupStatus{
			MirrorUUID:           C.GoString(gss.mirror_uuid),
			MirrorImageGlobalIDs: C.GoString(*gss.mirror_image_global_ids),
			MirrorImagePoolIDs:   int64(*gss.mirror_image_pool_ids),
			State:                MirrorGroupStatusState(gss.state),
			Description:          C.GoString(gss.description),
			MirrorImageCount:     int(gss.mirror_image_count),
			LastUpdate:           int64(gss.last_update),
			MirrorImages:         make([]SiteMirrorImageStatus, gss.mirror_image_count),
			Up:                   bool(gss.up),
		}

		sscs := (*siteArray)(unsafe.Pointer(gss.mirror_images))[:gss.mirror_image_count:gss.mirror_image_count]
		for i := C.uint32_t(0); i < gss.mirror_image_count; i++ {
			ss := sscs[i]
			status.SiteStatuses[i].MirrorImages[i] = SiteMirrorImageStatus{
				MirrorUUID:  C.GoString(ss.mirror_uuid),
				State:       MirrorImageStatusState(ss.state),
				Description: C.GoString(ss.description),
				LastUpdate:  int64(ss.last_update),
				Up:          bool(ss.up),
			}
		}
	}
	return status
}
