// +build !nautilus

// Initially, we're only providing mirroring related functions for octopus as
// that version of ceph deprecated a number of the functions in nautilus. If
// you need mirroring on an earlier supported version of ceph please file an
// issue in our tracker.

package rbd

// #cgo LDFLAGS: -lrbd
// #include <stdlib.h>
// #include <rbd/librbd.h>
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/ceph/go-ceph/internal/cutil"
	"github.com/ceph/go-ceph/internal/retry"
	"github.com/ceph/go-ceph/rados"
)

// MirrorMode is used to indicate an approach used for RBD mirroring.
type MirrorMode int64

const (
	// MirrorModeDisabled disables mirroring.
	MirrorModeDisabled = MirrorMode(C.RBD_MIRROR_MODE_DISABLED)
	// MirrorModeImage enables mirroring on a per-image basis.
	MirrorModeImage = MirrorMode(C.RBD_MIRROR_MODE_IMAGE)
	// MirrorModePool enables mirroring on all journaled images.
	MirrorModePool = MirrorMode(C.RBD_MIRROR_MODE_POOL)
)

// String representation of MirrorMode.
func (m MirrorMode) String() string {
	switch m {
	case MirrorModeDisabled:
		return "disabled"
	case MirrorModeImage:
		return "image"
	case MirrorModePool:
		return "pool"
	default:
		return "<unknown>"
	}
}

// ImageMirrorMode is used to indicate the mirroring approach for an RBD image.
type ImageMirrorMode int64

const (
	// ImageMirrorModeJournal uses journaling to propagate RBD images between
	// ceph clusters.
	ImageMirrorModeJournal = ImageMirrorMode(C.RBD_MIRROR_IMAGE_MODE_JOURNAL)
	// ImageMirrorModeSnapshot uses snapshots to propagate RBD images between
	// ceph clusters.
	ImageMirrorModeSnapshot = ImageMirrorMode(C.RBD_MIRROR_IMAGE_MODE_SNAPSHOT)
)

// String representation of ImageMirrorMode.
func (imm ImageMirrorMode) String() string {
	switch imm {
	case ImageMirrorModeJournal:
		return "journal"
	case ImageMirrorModeSnapshot:
		return "snapshot"
	default:
		return "<unknown>"
	}
}

// SetMirrorMode is used to enable or disable pool level mirroring with either
// an automatic or per-image behavior.
//
// Implements:
//  int rbd_mirror_mode_set(rados_ioctx_t io_ctx,
//                          rbd_mirror_mode_t mirror_mode);
func SetMirrorMode(ioctx *rados.IOContext, mode MirrorMode) error {
	ret := C.rbd_mirror_mode_set(
		cephIoctx(ioctx),
		C.rbd_mirror_mode_t(mode))
	return getError(ret)
}

// GetMirrorMode is used to fetch the current mirroring mode for a pool.
//
// Implements:
//  int rbd_mirror_mode_get(rados_ioctx_t io_ctx,
//                          rbd_mirror_mode_t *mirror_mode);
func GetMirrorMode(ioctx *rados.IOContext) (MirrorMode, error) {
	var mode C.rbd_mirror_mode_t

	ret := C.rbd_mirror_mode_get(
		cephIoctx(ioctx),
		&mode)
	if err := getError(ret); err != nil {
		return MirrorModeDisabled, err
	}
	return MirrorMode(mode), nil
}

// MirrorEnable will enable mirroring for an image using the specified mode.
//
// Implements:
//  int rbd_mirror_image_enable2(rbd_image_t image,
//                               rbd_mirror_image_mode_t mode);
func (image *Image) MirrorEnable(mode ImageMirrorMode) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}
	ret := C.rbd_mirror_image_enable2(image.image, C.rbd_mirror_image_mode_t(mode))
	return getError(ret)
}

// MirrorDisable will disable mirroring for the image.
//
// Implements:
//  int rbd_mirror_image_disable(rbd_image_t image, bool force);
func (image *Image) MirrorDisable(force bool) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}
	ret := C.rbd_mirror_image_disable(image.image, C.bool(force))
	return getError(ret)
}

// MirrorPromote will promote the image to primary status.
//
// Implements:
//  int rbd_mirror_image_promote(rbd_image_t image, bool force);
func (image *Image) MirrorPromote(force bool) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}
	ret := C.rbd_mirror_image_promote(image.image, C.bool(force))
	return getError(ret)
}

// MirrorDemote will demote the image to secondary status.
//
// Implements:
//  int rbd_mirror_image_demote(rbd_image_t image);
func (image *Image) MirrorDemote() error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}
	ret := C.rbd_mirror_image_demote(image.image)
	return getError(ret)
}

// MirrorResync is used to manually resolve split-brain status by triggering
// resynchronization.
//
// Implements:
//  int rbd_mirror_image_resync(rbd_image_t image);
func (image *Image) MirrorResync() error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}
	ret := C.rbd_mirror_image_resync(image.image)
	return getError(ret)
}

// MirrorInstanceID returns a string naming the instance id for the image.
//
// Implements:
//  int rbd_mirror_image_get_instance_id(rbd_image_t image,
//                                       char *instance_id,
//                                       size_t *id_max_length);
func (image *Image) MirrorInstanceID() (string, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return "", err
	}
	var (
		err   error
		buf   []byte
		cSize C.size_t
	)
	retry.WithSizes(1024, 1<<16, func(size int) retry.Hint {
		cSize = C.size_t(size)
		buf = make([]byte, cSize)
		ret := C.rbd_mirror_image_get_instance_id(
			image.image,
			(*C.char)(unsafe.Pointer(&buf[0])),
			&cSize)
		err = getErrorIfNegative(ret)
		return retry.Size(int(cSize)).If(err == errRange)
	})
	if err != nil {
		return "", err
	}
	return string(buf[:cSize]), nil
}

// MirrorImageState represents the mirroring state of a RBD image.
type MirrorImageState C.rbd_mirror_image_state_t

const (
	// MirrorImageDisabling is the representation of
	// RBD_MIRROR_IMAGE_DISABLING from librbd.
	MirrorImageDisabling = MirrorImageState(C.RBD_MIRROR_IMAGE_DISABLING)
	// MirrorImageEnabled is the representation of
	// RBD_MIRROR_IMAGE_ENABLED from librbd.
	MirrorImageEnabled = MirrorImageState(C.RBD_MIRROR_IMAGE_ENABLED)
	// MirrorImageDisabled is the representation of
	// RBD_MIRROR_IMAGE_DISABLED from librbd.
	MirrorImageDisabled = MirrorImageState(C.RBD_MIRROR_IMAGE_DISABLED)
)

// String representation of MirrorImageState.
func (mis MirrorImageState) String() string {
	switch mis {
	case MirrorImageDisabling:
		return "disabling"
	case MirrorImageEnabled:
		return "enabled"
	case MirrorImageDisabled:
		return "disabled"
	default:
		return "<unknown>"
	}
}

// MirrorImageInfo represents the mirroring status information of a RBD image.
type MirrorImageInfo struct {
	GlobalID string
	State    MirrorImageState
	Primary  bool
}

func convertMirrorImageInfo(cInfo *C.rbd_mirror_image_info_t) MirrorImageInfo {
	return MirrorImageInfo{
		GlobalID: C.GoString(cInfo.global_id),
		State:    MirrorImageState(cInfo.state),
		Primary:  bool(cInfo.primary),
	}
}

// GetMirrorImageInfo fetches the mirroring status information of a RBD image.
//
// Implements:
//  int rbd_mirror_image_get_info(rbd_image_t image,
//                                rbd_mirror_image_info_t *mirror_image_info,
//                                size_t info_size)
func (image *Image) GetMirrorImageInfo() (*MirrorImageInfo, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return nil, err
	}

	var cInfo C.rbd_mirror_image_info_t

	ret := C.rbd_mirror_image_get_info(
		image.image,
		&cInfo,
		C.sizeof_rbd_mirror_image_info_t)
	if ret < 0 {
		return nil, getError(ret)
	}

	mii := convertMirrorImageInfo(&cInfo)

	// free C memory allocated by C.rbd_mirror_image_get_info call
	C.rbd_mirror_image_get_info_cleanup(&cInfo)
	return &mii, nil
}

// GetImageMirrorMode fetches the mirroring approach for an RBD image.
//
// Implements:
//  int rbd_mirror_image_get_mode(rbd_image_t image, rbd_mirror_image_mode_t *mode);
func (image *Image) GetImageMirrorMode() (ImageMirrorMode, error) {
	var mode C.rbd_mirror_image_mode_t
	if err := image.validate(imageIsOpen); err != nil {
		return ImageMirrorMode(mode), err
	}

	ret := C.rbd_mirror_image_get_mode(image.image, &mode)
	return ImageMirrorMode(mode), getError(ret)
}

// MirrorImageStatusState is used to indicate the state of a mirrored image
// within the site status info.
type MirrorImageStatusState int64

const (
	// MirrorImageStatusStateUnknown is equivalent to MIRROR_IMAGE_STATUS_STATE_UNKNOWN
	MirrorImageStatusStateUnknown = MirrorImageStatusState(C.MIRROR_IMAGE_STATUS_STATE_UNKNOWN)
	// MirrorImageStatusStateError is equivalent to MIRROR_IMAGE_STATUS_STATE_ERROR
	MirrorImageStatusStateError = MirrorImageStatusState(C.MIRROR_IMAGE_STATUS_STATE_ERROR)
	// MirrorImageStatusStateSyncing is equivalent to MIRROR_IMAGE_STATUS_STATE_SYNCING
	MirrorImageStatusStateSyncing = MirrorImageStatusState(C.MIRROR_IMAGE_STATUS_STATE_SYNCING)
	// MirrorImageStatusStateStartingReplay is equivalent to MIRROR_IMAGE_STATUS_STATE_STARTING_REPLAY
	MirrorImageStatusStateStartingReplay = MirrorImageStatusState(C.MIRROR_IMAGE_STATUS_STATE_STARTING_REPLAY)
	// MirrorImageStatusStateReplaying is equivalent to MIRROR_IMAGE_STATUS_STATE_REPLAYING
	MirrorImageStatusStateReplaying = MirrorImageStatusState(C.MIRROR_IMAGE_STATUS_STATE_REPLAYING)
	// MirrorImageStatusStateStoppingReplay is equivalent to MIRROR_IMAGE_STATUS_STATE_STOPPING_REPLAY
	MirrorImageStatusStateStoppingReplay = MirrorImageStatusState(C.MIRROR_IMAGE_STATUS_STATE_STOPPING_REPLAY)
	// MirrorImageStatusStateStopped is equivalent to MIRROR_IMAGE_STATUS_STATE_STOPPED
	MirrorImageStatusStateStopped = MirrorImageStatusState(C.MIRROR_IMAGE_STATUS_STATE_STOPPED)
)

// String represents the MirrorImageStatusState as a short string.
func (state MirrorImageStatusState) String() (s string) {
	switch state {
	case MirrorImageStatusStateUnknown:
		s = "unknown"
	case MirrorImageStatusStateError:
		s = "error"
	case MirrorImageStatusStateSyncing:
		s = "syncing"
	case MirrorImageStatusStateStartingReplay:
		s = "starting_replay"
	case MirrorImageStatusStateReplaying:
		s = "replaying"
	case MirrorImageStatusStateStoppingReplay:
		s = "stopping_replay"
	case MirrorImageStatusStateStopped:
		s = "stopped"
	default:
		s = fmt.Sprintf("unknown(%d)", state)
	}
	return s
}

// SiteMirrorImageStatus contains information pertaining to the status of
// a mirrored image within a site.
type SiteMirrorImageStatus struct {
	MirrorUUID  string
	State       MirrorImageStatusState
	Description string
	LastUpdate  int64
	Up          bool
}

// GlobalMirrorImageStatus contains information pertaining to the global
// status of a mirrored image. It contains general information as well
// as per-site information stored in the SiteStatuses slice.
type GlobalMirrorImageStatus struct {
	Name         string
	Info         MirrorImageInfo
	SiteStatuses []SiteMirrorImageStatus
}

// LocalStatus returns one SiteMirrorImageStatus item from the SiteStatuses
// slice that corresponds to the local site's status. If the local status
// is not found than the error ErrNotExist will be returned.
func (gmis GlobalMirrorImageStatus) LocalStatus() (SiteMirrorImageStatus, error) {
	var (
		ss  SiteMirrorImageStatus
		err error = ErrNotExist
	)
	for i := range gmis.SiteStatuses {
		// I couldn't find it explicitly documented, but a site mirror uuid
		// of an empty string indicates that this is the local site.
		// This pattern occurs in both the pybind code and ceph c++.
		if gmis.SiteStatuses[i].MirrorUUID == "" {
			ss = gmis.SiteStatuses[i]
			err = nil
			break
		}
	}
	return ss, err
}

type siteArray [cutil.MaxIdx]C.rbd_mirror_image_site_status_t

// GetGlobalMirrorStatus returns status information pertaining to the state
// of the images's mirroring.
//
// Implements:
//   int rbd_mirror_image_get_global_status(
//     rbd_image_t image,
//     rbd_mirror_image_global_status_t *mirror_image_global_status,
//     size_t status_size);
func (image *Image) GetGlobalMirrorStatus() (GlobalMirrorImageStatus, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return GlobalMirrorImageStatus{}, err
	}

	s := C.rbd_mirror_image_global_status_t{}
	ret := C.rbd_mirror_image_get_global_status(
		image.image,
		&s,
		C.sizeof_rbd_mirror_image_global_status_t)
	if err := getError(ret); err != nil {
		return GlobalMirrorImageStatus{}, err
	}
	defer C.rbd_mirror_image_global_status_cleanup(&s)

	status := GlobalMirrorImageStatus{
		Name:         C.GoString(s.name),
		Info:         convertMirrorImageInfo(&s.info),
		SiteStatuses: make([]SiteMirrorImageStatus, s.site_statuses_count),
	}
	// use the "Sven Technique" to treat the C pointer as a go slice temporarily
	sscs := (*siteArray)(unsafe.Pointer(s.site_statuses))[:s.site_statuses_count:s.site_statuses_count]
	for i := C.uint32_t(0); i < s.site_statuses_count; i++ {
		ss := sscs[i]
		status.SiteStatuses[i] = SiteMirrorImageStatus{
			MirrorUUID:  C.GoString(ss.mirror_uuid),
			State:       MirrorImageStatusState(ss.state),
			Description: C.GoString(ss.description),
			LastUpdate:  int64(ss.last_update),
			Up:          bool(ss.up),
		}
	}
	return status, nil
}

// CreateMirrorSnapshot creates a snapshot for image propagation to mirrors.
//
// Implements:
//  int rbd_mirror_image_create_snapshot(rbd_image_t image,
//                                       uint64_t *snap_id);
func (image *Image) CreateMirrorSnapshot() (uint64, error) {
	var snapID C.uint64_t
	ret := C.rbd_mirror_image_create_snapshot(
		image.image,
		&snapID)
	return uint64(snapID), getError(ret)
}

// MirrorImageStatusSummary returns a map of images statuses and the count
// of images with said status.
//
// Implements:
//  int rbd_mirror_image_status_summary(
//    rados_ioctx_t io_ctx, rbd_mirror_image_status_state_t *states, int *counts,
//    size_t *maxlen);
func MirrorImageStatusSummary(
	ioctx *rados.IOContext) (map[MirrorImageStatusState]uint, error) {
	// ideally, we already know the size of the arrays - they should be
	// the size of all the values of the rbd_mirror_image_status_state_t
	// enum. But the C api doesn't enforce this so we give a little
	// wiggle room in case the server returns values outside the enum
	// we know about. This is the only case I can think of that we'd
	// be able to get -ERANGE.
	var (
		cioctx  = cephIoctx(ioctx)
		err     error
		cStates []C.rbd_mirror_image_status_state_t
		cCounts []C.int
		cSize   C.size_t
	)
	retry.WithSizes(16, 1<<16, func(size int) retry.Hint {
		cSize = C.size_t(size)
		cStates = make([]C.rbd_mirror_image_status_state_t, cSize)
		cCounts = make([]C.int, cSize)
		ret := C.rbd_mirror_image_status_summary(
			cioctx,
			(*C.rbd_mirror_image_status_state_t)(&cStates[0]),
			(*C.int)(&cCounts[0]),
			&cSize)
		err = getErrorIfNegative(ret)
		return retry.Size(int(cSize)).If(err == errRange)
	})
	if err != nil {
		return nil, err
	}

	m := map[MirrorImageStatusState]uint{}
	for i := 0; i < int(cSize); i++ {
		s := MirrorImageStatusState(cStates[i])
		m[s] = uint(cCounts[i])
	}
	return m, nil
}
