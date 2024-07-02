package rbd

/*
#cgo LDFLAGS: -lrbd
#include <stdlib.h>
#include <rbd/librbd.h>
*/
import "C"

import (
	"unsafe"

	"github.com/ceph/go-ceph/internal/cutil"
	"github.com/ceph/go-ceph/internal/retry"
	"github.com/ceph/go-ceph/rados"
)

// GroupCreate is used to create an image group.
//
// Implements:
//
//	int rbd_group_create(rados_ioctx_t p, const char *name);
func GroupCreate(ioctx *rados.IOContext, name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.rbd_group_create(cephIoctx(ioctx), cName)
	return getError(ret)
}

// GroupRemove is used to remove an image group.
//
// Implements:
//
//	int rbd_group_remove(rados_ioctx_t p, const char *name);
func GroupRemove(ioctx *rados.IOContext, name string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	ret := C.rbd_group_remove(cephIoctx(ioctx), cName)
	return getError(ret)
}

// GroupRename will rename an existing image group.
//
// Implements:
//
//	int rbd_group_rename(rados_ioctx_t p, const char *src_name,
//	                     const char *dest_name);
func GroupRename(ioctx *rados.IOContext, src, dest string) error {
	cSrc := C.CString(src)
	defer C.free(unsafe.Pointer(cSrc))
	cDest := C.CString(dest)
	defer C.free(unsafe.Pointer(cDest))

	ret := C.rbd_group_rename(cephIoctx(ioctx), cSrc, cDest)
	return getError(ret)
}

// GroupList returns a slice of image group names.
//
// Implements:
//
//	int rbd_group_list(rados_ioctx_t p, char *names, size_t *size);
func GroupList(ioctx *rados.IOContext) ([]string, error) {
	var (
		buf []byte
		err error
		ret C.int
	)
	retry.WithSizes(1024, 262144, func(size int) retry.Hint {
		cSize := C.size_t(size)
		buf = make([]byte, cSize)
		ret = C.rbd_group_list(
			cephIoctx(ioctx),
			(*C.char)(unsafe.Pointer(&buf[0])),
			&cSize)
		err = getErrorIfNegative(ret)
		return retry.Size(int(cSize)).If(err == errRange)
	})

	if err != nil {
		return nil, err
	}

	// cSize is not set to the expected size when it is sufficiently large
	// but ret will be set to the size in a non-error condition.
	groups := cutil.SplitBuffer(buf[:ret])
	return groups, nil
}

// GroupImageAdd will add the specified image to the named group.
// An io context must be supplied for both the group and image.
//
// Implements:
//
//	int rbd_group_image_add(rados_ioctx_t group_p,
//	                        const char *group_name,
//	                        rados_ioctx_t image_p,
//	                        const char *image_name);
func GroupImageAdd(groupIoctx *rados.IOContext, groupName string,
	imageIoctx *rados.IOContext, imageName string) error {

	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))
	cImageName := C.CString(imageName)
	defer C.free(unsafe.Pointer(cImageName))

	ret := C.rbd_group_image_add(
		cephIoctx(groupIoctx),
		cGroupName,
		cephIoctx(imageIoctx),
		cImageName,
		C.uint32_t(0))
	return getError(ret)
}

// GroupImageRemove will remove the specified image from the named group.
// An io context must be supplied for both the group and image.
//
// Implements:
//
//	int rbd_group_image_remove(rados_ioctx_t group_p,
//	                           const char *group_name,
//	                           rados_ioctx_t image_p,
//	                           const char *image_name);
func GroupImageRemove(groupIoctx *rados.IOContext, groupName string,
	imageIoctx *rados.IOContext, imageName string) error {

	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))
	cImageName := C.CString(imageName)
	defer C.free(unsafe.Pointer(cImageName))

	ret := C.rbd_group_image_remove(
		cephIoctx(groupIoctx),
		cGroupName,
		cephIoctx(imageIoctx),
		cImageName,
		C.uint32_t(0))
	return getError(ret)
}

// GroupImageRemoveByID will remove the specified image from the named group.
// An io context must be supplied for both the group and image.
//
// Implements:
//
//	CEPH_RBD_API int rbd_group_image_remove_by_id(rados_ioctx_t group_p,
//	                                             const char *group_name,
//	                                             rados_ioctx_t image_p,
//	                                             const char *image_id);
func GroupImageRemoveByID(groupIoctx *rados.IOContext, groupName string,
	imageIoctx *rados.IOContext, imageID string) error {

	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))
	cid := C.CString(imageID)
	defer C.free(unsafe.Pointer(cid))

	ret := C.rbd_group_image_remove_by_id(
		cephIoctx(groupIoctx),
		cGroupName,
		cephIoctx(imageIoctx),
		cid,
		C.uint32_t(0))
	return getError(ret)
}

// GroupImageState indicates an image's state in a group.
type GroupImageState int

const (
	// GroupImageStateAttached is equivalent to RBD_GROUP_IMAGE_STATE_ATTACHED
	GroupImageStateAttached = GroupImageState(C.RBD_GROUP_IMAGE_STATE_ATTACHED)
	// GroupImageStateIncomplete is equivalent to RBD_GROUP_IMAGE_STATE_INCOMPLETE
	GroupImageStateIncomplete = GroupImageState(C.RBD_GROUP_IMAGE_STATE_INCOMPLETE)
)

// GroupImageInfo reports on images within a group.
type GroupImageInfo struct {
	Name   string
	PoolID int64
	State  GroupImageState
}

// GroupImageList returns a slice of GroupImageInfo types based on the
// images that are part of the named group.
//
// Implements:
//
//	int rbd_group_image_list(rados_ioctx_t group_p,
//	                         const char *group_name,
//	                         rbd_group_image_info_t *images,
//	                         size_t group_image_info_size,
//	                         size_t *num_entries);
func GroupImageList(ioctx *rados.IOContext, name string) ([]GroupImageInfo, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var (
		cImages []C.rbd_group_image_info_t
		cSize   C.size_t
		err     error
	)
	retry.WithSizes(1024, 262144, func(size int) retry.Hint {
		cSize = C.size_t(size)
		cImages = make([]C.rbd_group_image_info_t, cSize)
		ret := C.rbd_group_image_list(
			cephIoctx(ioctx),
			cName,
			(*C.rbd_group_image_info_t)(unsafe.Pointer(&cImages[0])),
			C.sizeof_rbd_group_image_info_t,
			&cSize)
		err = getErrorIfNegative(ret)
		return retry.Size(int(cSize)).If(err == errRange)
	})

	if err != nil {
		return nil, err
	}

	images := make([]GroupImageInfo, cSize)
	for i := range images {
		images[i].Name = C.GoString(cImages[i].name)
		images[i].PoolID = int64(cImages[i].pool)
		images[i].State = GroupImageState(cImages[i].state)
	}

	// free C memory allocated by C.rbd_group_image_list call
	ret := C.rbd_group_image_list_cleanup(
		(*C.rbd_group_image_info_t)(unsafe.Pointer(&cImages[0])),
		C.sizeof_rbd_group_image_info_t,
		cSize)
	return images, getError(ret)
}

// GroupInfo contains the name and pool id of a RBD group.
type GroupInfo struct {
	Name   string
	PoolID int64
}

// GetGroup returns group info for the group this image is part of.
//
// Implements:
//
//	int rbd_get_group(rbd_image_t image, rbd_group_info_t *group_info,
//	                  size_t group_info_size);
func (image *Image) GetGroup() (GroupInfo, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return GroupInfo{}, err
	}

	var cgi C.rbd_group_info_t
	ret := C.rbd_get_group(
		image.image,
		&cgi,
		C.sizeof_rbd_group_info_t)
	if err := getErrorIfNegative(ret); err != nil {
		return GroupInfo{}, err
	}

	gi := GroupInfo{
		Name:   C.GoString(cgi.name),
		PoolID: int64(cgi.pool),
	}
	ret = C.rbd_group_info_cleanup(&cgi, C.sizeof_rbd_group_info_t)
	return gi, getError(ret)
}

// MirrorGroupStatusState is used to indicate the state of a mirrored group
// within the site status info.
type MirrorGroupStatusState int64

const (
	// MirrorGrouptatusStateUnknown is equivalent to MIRROR_GROUP_STATUS_STATE_UNKNOWN
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

// MirrorImageState represents the mirroring state of a RBD image.
type MirrorGroupState C.rbd_mirror_group_state_t

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

// SiteMirrorGroupStatus contains information pertaining to the status of
// a mirrored group within a site.
type SiteMirrorGroupStatus struct {
	MirrorUUID           string
	State                MirrorGroupStatusState
	MirrorImageCount     int
	MirrorImagePoolIds   int64
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
//		IoCtx& io_ctx,
//		const char *group_name
//		mirror_group_global_status_t *mirror_group_status,
//		size_t status_size);
func GetGlobalMirrorGroupStatus(ioctx *rados.IOContext, groupName string) (GlobalMirrorGroupStatus, error) {
	s := C.rbd_mirror_group_global_status_t{}
	cGroupName := C.CString(groupName)
	defer C.free(unsafe.Pointer(cGroupName))
	ret := C.rbd_mirror_group_get_global_status(
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
			MirrorUUID:       C.GoString(gss.mirror_uuid),
			State:            MirrorGroupStatusState(gss.state),
			Description:      C.GoString(gss.description),
			MirrorImageCount: int(gss.mirror_image_count),
			LastUpdate:       int64(gss.last_update),
			MirrorImages:     make([]SiteMirrorImageStatus, gss.mirror_image_count),
			Up:               bool(gss.up),
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

func convertMirrorGroupInfo(cInfo *C.rbd_mirror_group_info_t) MirrorGroupInfo {
	return MirrorGroupInfo{
		GlobalID:        C.GoString(cInfo.global_id),
		MirrorImageMode: ImageMirrorMode(cInfo.mirror_image_mode),
		State:           MirrorGroupState(cInfo.state),
		Primary:         bool(cInfo.primary),
	}
}
