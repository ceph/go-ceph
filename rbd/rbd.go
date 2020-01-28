package rbd

// #cgo LDFLAGS: -lrbd
// /* force XSI-complaint strerror_r() */
// #define _POSIX_C_SOURCE 200112L
// #undef _GNU_SOURCE
// #include <errno.h>
// #include <stdlib.h>
// #include <rados/librados.h>
// #include <rbd/librbd.h>
// #include <rbd/features.h>
import "C"

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"time"
	"unsafe"

	"github.com/ceph/go-ceph/errutil"
	"github.com/ceph/go-ceph/rados"
)

const (
	// RBD features.
	RbdFeatureLayering      = uint64(C.RBD_FEATURE_LAYERING)
	RbdFeatureStripingV2    = uint64(C.RBD_FEATURE_STRIPINGV2)
	RbdFeatureExclusiveLock = uint64(C.RBD_FEATURE_EXCLUSIVE_LOCK)
	RbdFeatureObjectMap     = uint64(C.RBD_FEATURE_OBJECT_MAP)
	RbdFeatureFastDiff      = uint64(C.RBD_FEATURE_FAST_DIFF)
	RbdFeatureDeepFlatten   = uint64(C.RBD_FEATURE_DEEP_FLATTEN)
	RbdFeatureJournaling    = uint64(C.RBD_FEATURE_JOURNALING)
	RbdFeatureDataPool      = uint64(C.RBD_FEATURE_DATA_POOL)

	RbdFeaturesDefault = uint64(C.RBD_FEATURES_DEFAULT)

	// Features that make an image inaccessible for read or write by clients that don't understand
	// them.
	RbdFeaturesIncompatible = uint64(C.RBD_FEATURES_INCOMPATIBLE)

	// Features that make an image unwritable by clients that don't understand them.
	RbdFeaturesRwIncompatible = uint64(C.RBD_FEATURES_RW_INCOMPATIBLE)

	// Features that may be dynamically enabled or disabled.
	RbdFeaturesMutable = uint64(C.RBD_FEATURES_MUTABLE)

	// Features that only work when used with a single client using the image for writes.
	RbdFeaturesSingleClient = uint64(C.RBD_FEATURES_SINGLE_CLIENT)

	// Image.Seek() constants
	SeekSet = int(C.SEEK_SET)
	SeekCur = int(C.SEEK_CUR)
	SeekEnd = int(C.SEEK_END)
)

// bits for Image.validate() and Snapshot.validate()
const (
	imageNeedsName uint32 = 1 << iota
	imageNeedsIOContext
	imageIsOpen
	snapshotNeedsName

	// NoSnapshot indicates that no snapshot name is in use (see OpenImage)
	NoSnapshot = ""
)

//
type RBDError int

var (
	ErrNoIOContext    = errors.New("RBD image does not have an IOContext")
	ErrNoName         = errors.New("RBD image does not have a name")
	ErrSnapshotNoName = errors.New("RBD snapshot does not have a name")
	ErrImageNotOpen   = errors.New("RBD image not open")
	ErrNotFound       = errors.New("RBD image not found")

	// retained for compatibility with old versions
	RbdErrorImageNotOpen = ErrImageNotOpen
	RbdErrorNotFound     = ErrNotFound
)

//
type ImageInfo struct {
	Size              uint64
	Obj_size          uint64
	Num_objs          uint64
	Order             int
	Block_name_prefix string
	Parent_pool       int64
	Parent_name       string
}

//
type SnapInfo struct {
	Id   uint64
	Size uint64
	Name string
}

//
type Locker struct {
	Client string
	Cookie string
	Addr   string
}

//
type Image struct {
	io.Reader
	io.Writer
	io.Seeker
	io.ReaderAt
	io.WriterAt
	name   string
	offset int64
	ioctx  *rados.IOContext
	image  C.rbd_image_t
}

//
type Snapshot struct {
	image *Image
	name  string
}

// TrashInfo contains information about trashed RBDs.
type TrashInfo struct {
	Id               string    // Id string, required to remove / restore trashed RBDs.
	Name             string    // Original name of trashed RBD.
	DeletionTime     time.Time // Date / time at which the RBD was moved to the trash.
	DefermentEndTime time.Time // Date / time after which the trashed RBD may be permanently deleted.
}

//
func split(buf []byte) (values []string) {
	tmp := bytes.Split(buf[:len(buf)-1], []byte{0})
	for _, s := range tmp {
		if len(s) > 0 {
			go_s := C.GoString((*C.char)(unsafe.Pointer(&s[0])))
			values = append(values, go_s)
		}
	}
	return values
}

// test if a bit is set in the given value
func hasBit(value, bit uint32) bool {
	return (value & bit) == bit
}

// validate the attributes listed in the req bitmask, and return an error in
// case the attribute is not set
func (image *Image) validate(req uint32) error {
	if hasBit(req, imageNeedsName) && image.name == "" {
		return ErrNoName
	} else if hasBit(req, imageNeedsIOContext) && image.ioctx == nil {
		return ErrNoIOContext
	} else if hasBit(req, imageIsOpen) && image.image == nil {
		return ErrImageNotOpen
	}

	return nil
}

// validate the attributes listed in the req bitmask, and return an error in
// case the attribute is not set
// Calls snapshot.image.validate(req) to validate the image attributes.
func (snapshot *Snapshot) validate(req uint32) error {
	if hasBit(req, snapshotNeedsName) && snapshot.name == "" {
		return ErrSnapshotNoName
	} else if snapshot.image != nil {
		return snapshot.image.validate(req)
	}

	return nil
}

func (e RBDError) Error() string {
	errno, s := errutil.FormatErrno(int(e))
	if s == "" {
		return fmt.Sprintf("rbd: ret=%d", errno)
	}
	return fmt.Sprintf("rbd: ret=%d, %s", errno, s)
}

func getError(err C.int) error {
	if err != 0 {
		if err == -C.ENOENT {
			return ErrNotFound
		}
		return RBDError(err)
	} else {
		return nil
	}
}

// Version returns the major, minor, and patch level of the librbd library.
func Version() (int, int, int) {
	var c_major, c_minor, c_patch C.int
	C.rbd_version(&c_major, &c_minor, &c_patch)
	return int(c_major), int(c_minor), int(c_patch)
}

// GetImage gets a reference to a previously created rbd image.
func GetImage(ioctx *rados.IOContext, name string) *Image {
	return &Image{
		ioctx: ioctx,
		name:  name,
	}
}

// Create a new rbd image.
//
// Implements:
//  int rbd_create(rados_ioctx_t io, const char *name, uint64_t size, int *order);
//
// Also implements (for backward compatibility):
//  int rbd_create2(rados_ioctx_t io, const char *name, uint64_t size,
//          uint64_t features, int *order);
//  int rbd_create3(rados_ioctx_t io, const char *name, uint64_t size,
//        uint64_t features, int *order,
//        uint64_t stripe_unit, uint64_t stripe_count);
func Create(ioctx *rados.IOContext, name string, size uint64, order int,
	args ...uint64) (image *Image, err error) {
	var ret C.int

	switch len(args) {
	case 3:
		return Create3(ioctx, name, size, args[0], order, args[1],
			args[2])
	case 1:
		return Create2(ioctx, name, size, args[0], order)
	case 0:
		c_order := C.int(order)
		c_name := C.CString(name)

		defer C.free(unsafe.Pointer(c_name))

		ret = C.rbd_create(C.rados_ioctx_t(ioctx.Pointer()),
			c_name, C.uint64_t(size), &c_order)
	default:
		return nil, errors.New("Wrong number of argument")
	}

	if ret < 0 {
		return nil, RBDError(ret)
	}

	return &Image{
		ioctx: ioctx,
		name:  name,
	}, nil
}

// Create2 creates a new rbd image using provided features.
//
// Implements:
//  int rbd_create2(rados_ioctx_t io, const char *name, uint64_t size,
//          uint64_t features, int *order);
func Create2(ioctx *rados.IOContext, name string, size uint64, features uint64,
	order int) (image *Image, err error) {
	var ret C.int

	c_order := C.int(order)
	c_name := C.CString(name)

	defer C.free(unsafe.Pointer(c_name))

	ret = C.rbd_create2(C.rados_ioctx_t(ioctx.Pointer()), c_name,
		C.uint64_t(size), C.uint64_t(features), &c_order)
	if ret < 0 {
		return nil, RBDError(ret)
	}

	return &Image{
		ioctx: ioctx,
		name:  name,
	}, nil
}

// Create3 creates a new rbd image using provided features and stripe
// parameters.
//
// Implements:
//  int rbd_create3(rados_ioctx_t io, const char *name, uint64_t size,
//        uint64_t features, int *order,
//        uint64_t stripe_unit, uint64_t stripe_count);
func Create3(ioctx *rados.IOContext, name string, size uint64, features uint64,
	order int, stripe_unit uint64, stripe_count uint64) (image *Image, err error) {
	var ret C.int

	c_order := C.int(order)
	c_name := C.CString(name)

	defer C.free(unsafe.Pointer(c_name))

	ret = C.rbd_create3(C.rados_ioctx_t(ioctx.Pointer()), c_name,
		C.uint64_t(size), C.uint64_t(features), &c_order,
		C.uint64_t(stripe_unit), C.uint64_t(stripe_count))
	if ret < 0 {
		return nil, RBDError(ret)
	}

	return &Image{
		ioctx: ioctx,
		name:  name,
	}, nil
}

// Clone a new rbd image from a snapshot.
//
// Implements:
//  int rbd_clone(rados_ioctx_t p_ioctx, const char *p_name,
//           const char *p_snapname, rados_ioctx_t c_ioctx,
//           const char *c_name, uint64_t features, int *c_order);
func (image *Image) Clone(snapname string, c_ioctx *rados.IOContext, c_name string, features uint64, order int) (*Image, error) {
	if err := image.validate(imageNeedsIOContext); err != nil {
		return nil, err
	}

	c_order := C.int(order)
	c_p_name := C.CString(image.name)
	c_p_snapname := C.CString(snapname)
	c_c_name := C.CString(c_name)

	defer C.free(unsafe.Pointer(c_p_name))
	defer C.free(unsafe.Pointer(c_p_snapname))
	defer C.free(unsafe.Pointer(c_c_name))

	ret := C.rbd_clone(C.rados_ioctx_t(image.ioctx.Pointer()),
		c_p_name, c_p_snapname,
		C.rados_ioctx_t(c_ioctx.Pointer()),
		c_c_name, C.uint64_t(features), &c_order)
	if ret < 0 {
		return nil, RBDError(ret)
	}

	return &Image{
		ioctx: c_ioctx,
		name:  c_name,
	}, nil
}

// Remove the specified rbd image.
//
// Implements:
//  int rbd_remove(rados_ioctx_t io, const char *name);
func (image *Image) Remove() error {
	if err := image.validate(imageNeedsIOContext | imageNeedsName); err != nil {
		return err
	}
	return RemoveImage(image.ioctx, image.name)
}

// Trash will move an image into the RBD trash, where it will be protected (i.e., salvageable) for
// at least the specified delay.
func (image *Image) Trash(delay time.Duration) error {
	if err := image.validate(imageNeedsIOContext | imageNeedsName); err != nil {
		return err
	}

	c_name := C.CString(image.name)
	defer C.free(unsafe.Pointer(c_name))

	return getError(C.rbd_trash_move(C.rados_ioctx_t(image.ioctx.Pointer()), c_name,
		C.uint64_t(delay.Seconds())))
}

// Rename an rbd image.
//
// Implements:
//  int rbd_rename(rados_ioctx_t src_io_ctx, const char *srcname, const char *destname);
func (image *Image) Rename(destname string) error {
	if err := image.validate(imageNeedsIOContext | imageNeedsName); err != nil {
		return err
	}

	c_srcname := C.CString(image.name)
	c_destname := C.CString(destname)

	defer C.free(unsafe.Pointer(c_srcname))
	defer C.free(unsafe.Pointer(c_destname))

	err := RBDError(C.rbd_rename(C.rados_ioctx_t(image.ioctx.Pointer()),
		c_srcname, c_destname))
	if err == 0 {
		image.name = destname
		return nil
	}
	return err
}

// Open the rbd image.
//
// Deprecated: The Open function was provided in earlier versions of the API
// and now exists to support older code. The use of OpenImage and
// OpenImageReadOnly is preferred.
func (image *Image) Open(args ...interface{}) error {
	if err := image.validate(imageNeedsIOContext | imageNeedsName); err != nil {
		return err
	}

	var (
		snapName string
		readOnly bool
	)
	for _, arg := range args {
		switch t := arg.(type) {
		case string:
			snapName = t
		case bool:
			readOnly = t
		default:
			return errors.New("Unexpected argument")
		}
	}

	var (
		tmp *Image
		err error
	)
	if readOnly {
		tmp, err = OpenImageReadOnly(image.ioctx, image.name, snapName)
	} else {
		tmp, err = OpenImage(image.ioctx, image.name, snapName)
	}
	if err != nil {
		return err
	}

	image.image = tmp.image
	return nil
}

// Close an open rbd image.
//
// Implements:
//  int rbd_close(rbd_image_t image);
func (image *Image) Close() error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	if ret := C.rbd_close(image.image); ret != 0 {
		return RBDError(ret)
	}

	image.image = nil
	return nil
}

// Resize an rbd image.
//
// Implements:
//  int rbd_resize(rbd_image_t image, uint64_t size);
func (image *Image) Resize(size uint64) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	return getError(C.rbd_resize(image.image, C.uint64_t(size)))
}

// Stat an rbd image.
//
// Implements:
//  int rbd_stat(rbd_image_t image, rbd_image_info_t *info, size_t infosize);
func (image *Image) Stat() (info *ImageInfo, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return nil, err
	}

	var c_stat C.rbd_image_info_t

	if ret := C.rbd_stat(image.image, &c_stat, C.size_t(unsafe.Sizeof(info))); ret < 0 {
		return info, RBDError(ret)
	}

	return &ImageInfo{
		Size:              uint64(c_stat.size),
		Obj_size:          uint64(c_stat.obj_size),
		Num_objs:          uint64(c_stat.num_objs),
		Order:             int(c_stat.order),
		Block_name_prefix: C.GoString((*C.char)(&c_stat.block_name_prefix[0])),
		Parent_pool:       int64(c_stat.parent_pool),
		Parent_name:       C.GoString((*C.char)(&c_stat.parent_name[0]))}, nil
}

// IsOldFormat returns true if the rbd image uses the old format.
//
// Implements:
//  int rbd_get_old_format(rbd_image_t image, uint8_t *old);
func (image *Image) IsOldFormat() (old_format bool, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return false, err
	}

	var c_old_format C.uint8_t
	ret := C.rbd_get_old_format(image.image,
		&c_old_format)
	if ret < 0 {
		return false, RBDError(ret)
	}

	return c_old_format != 0, nil
}

// GetSize returns the size of the rbd image.
//
// Implements:
//  int rbd_size(rbd_image_t image, uint64_t *size);
func (image *Image) GetSize() (size uint64, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if ret := C.rbd_get_size(image.image, (*C.uint64_t)(&size)); ret < 0 {
		return 0, RBDError(ret)
	}

	return size, nil
}

// GetFeatures returns the features bitmask for the rbd image.
//
// Implements:
//  int rbd_get_features(rbd_image_t image, uint64_t *features);
func (image *Image) GetFeatures() (features uint64, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if ret := C.rbd_get_features(image.image, (*C.uint64_t)(&features)); ret < 0 {
		return 0, RBDError(ret)
	}

	return features, nil
}

// GetStripeUnit returns the stripe-unit value for the rbd image.
//
// Implements:
//  int rbd_get_stripe_unit(rbd_image_t image, uint64_t *stripe_unit);
func (image *Image) GetStripeUnit() (stripe_unit uint64, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if ret := C.rbd_get_stripe_unit(image.image, (*C.uint64_t)(&stripe_unit)); ret < 0 {
		return 0, RBDError(ret)
	}

	return stripe_unit, nil
}

// GetStripeCount returns the stripe-count value for the rbd image.
//
// Implements:
//  int rbd_get_stripe_count(rbd_image_t image, uint64_t *stripe_count);
func (image *Image) GetStripeCount() (stripe_count uint64, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if ret := C.rbd_get_stripe_count(image.image, (*C.uint64_t)(&stripe_count)); ret < 0 {
		return 0, RBDError(ret)
	}

	return stripe_count, nil
}

// GetOverlap returns the overlapping bytes between the rbd image and its
// parent.
//
// Implements:
//  int rbd_get_overlap(rbd_image_t image, uint64_t *overlap);
func (image *Image) GetOverlap() (overlap uint64, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if ret := C.rbd_get_overlap(image.image, (*C.uint64_t)(&overlap)); ret < 0 {
		return overlap, RBDError(ret)
	}

	return overlap, nil
}

// Copy one rbd image to another.
//
// Implements:
//  int rbd_copy(rbd_image_t image, rados_ioctx_t dest_io_ctx, const char *destname);
func (image *Image) Copy(ioctx *rados.IOContext, destname string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	} else if ioctx == nil {
		return ErrNoIOContext
	} else if len(destname) == 0 {
		return ErrNoName
	}

	c_destname := C.CString(destname)
	defer C.free(unsafe.Pointer(c_destname))

	return getError(C.rbd_copy(image.image,
		C.rados_ioctx_t(ioctx.Pointer()), c_destname))
}

// Copy one rbd image to another, using an image handle.
//
// Implements:
//  int rbd_copy2(rbd_image_t src, rbd_image_t dest);
func (image *Image) Copy2(dest *Image) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	} else if err := dest.validate(imageIsOpen); err != nil {
		return err
	}

	return getError(C.rbd_copy2(image.image, dest.image))
}

// Flatten removes snapshot references from the image.
//
// Implements:
//  int rbd_flatten(rbd_image_t image);
func (image *Image) Flatten() error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	return getError(C.rbd_flatten(image.image))
}

// ListChildren returns a list of images that reference the current snapshot.
//
// Implements:
//  ssize_t rbd_list_children(rbd_image_t image, char *pools, size_t *pools_len,
//               char *images, size_t *images_len);
func (image *Image) ListChildren() (pools []string, images []string, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return nil, nil, err
	}

	var c_pools_len, c_images_len C.size_t

	ret := C.rbd_list_children(image.image,
		nil, &c_pools_len,
		nil, &c_images_len)
	if ret == 0 {
		return nil, nil, nil
	}
	if ret < 0 && ret != -C.ERANGE {
		return nil, nil, RBDError(ret)
	}

	pools_buf := make([]byte, c_pools_len)
	images_buf := make([]byte, c_images_len)

	ret = C.rbd_list_children(image.image,
		(*C.char)(unsafe.Pointer(&pools_buf[0])),
		&c_pools_len,
		(*C.char)(unsafe.Pointer(&images_buf[0])),
		&c_images_len)
	if ret < 0 {
		return nil, nil, RBDError(ret)
	}

	tmp := bytes.Split(pools_buf[:c_pools_len-1], []byte{0})
	for _, s := range tmp {
		if len(s) > 0 {
			name := C.GoString((*C.char)(unsafe.Pointer(&s[0])))
			pools = append(pools, name)
		}
	}

	tmp = bytes.Split(images_buf[:c_images_len-1], []byte{0})
	for _, s := range tmp {
		if len(s) > 0 {
			name := C.GoString((*C.char)(unsafe.Pointer(&s[0])))
			images = append(images, name)
		}
	}

	return pools, images, nil
}

// ListLockers returns a list of clients that have locks on the image.
//
// Impelemnts:
//  ssize_t rbd_list_lockers(rbd_image_t image, int *exclusive,
//              char *tag, size_t *tag_len,
//              char *clients, size_t *clients_len,
//              char *cookies, size_t *cookies_len,
//              char *addrs, size_t *addrs_len);
func (image *Image) ListLockers() (tag string, lockers []Locker, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return "", nil, err
	}

	var c_exclusive C.int
	var c_tag_len, c_clients_len, c_cookies_len, c_addrs_len C.size_t
	var c_locker_cnt C.ssize_t

	C.rbd_list_lockers(image.image, &c_exclusive,
		nil, (*C.size_t)(&c_tag_len),
		nil, (*C.size_t)(&c_clients_len),
		nil, (*C.size_t)(&c_cookies_len),
		nil, (*C.size_t)(&c_addrs_len))

	// no locker held on rbd image when either c_clients_len,
	// c_cookies_len or c_addrs_len is *0*, so just quickly returned
	if int(c_clients_len) == 0 || int(c_cookies_len) == 0 ||
		int(c_addrs_len) == 0 {
		lockers = make([]Locker, 0)
		return "", lockers, nil
	}

	tag_buf := make([]byte, c_tag_len)
	clients_buf := make([]byte, c_clients_len)
	cookies_buf := make([]byte, c_cookies_len)
	addrs_buf := make([]byte, c_addrs_len)

	c_locker_cnt = C.rbd_list_lockers(image.image, &c_exclusive,
		(*C.char)(unsafe.Pointer(&tag_buf[0])), (*C.size_t)(&c_tag_len),
		(*C.char)(unsafe.Pointer(&clients_buf[0])), (*C.size_t)(&c_clients_len),
		(*C.char)(unsafe.Pointer(&cookies_buf[0])), (*C.size_t)(&c_cookies_len),
		(*C.char)(unsafe.Pointer(&addrs_buf[0])), (*C.size_t)(&c_addrs_len))

	// rbd_list_lockers returns negative value for errors
	// and *0* means no locker held on rbd image.
	// but *0* is unexpected here because first rbd_list_lockers already
	// dealt with no locker case
	if int(c_locker_cnt) <= 0 {
		return "", nil, RBDError(c_locker_cnt)
	}

	clients := split(clients_buf)
	cookies := split(cookies_buf)
	addrs := split(addrs_buf)

	lockers = make([]Locker, c_locker_cnt)
	for i := 0; i < int(c_locker_cnt); i++ {
		lockers[i] = Locker{Client: clients[i],
			Cookie: cookies[i],
			Addr:   addrs[i]}
	}

	return string(tag_buf), lockers, nil
}

// LockExclusive acquires an exclusive lock on the rbd image.
//
// Implements:
//  int rbd_lock_exclusive(rbd_image_t image, const char *cookie);
func (image *Image) LockExclusive(cookie string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_cookie := C.CString(cookie)
	defer C.free(unsafe.Pointer(c_cookie))

	return getError(C.rbd_lock_exclusive(image.image, c_cookie))
}

// LockShared acquires a shared lock on the rbd image.
//
// Implements:
//  int rbd_lock_shared(rbd_image_t image, const char *cookie, const char *tag);
func (image *Image) LockShared(cookie string, tag string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_cookie := C.CString(cookie)
	c_tag := C.CString(tag)
	defer C.free(unsafe.Pointer(c_cookie))
	defer C.free(unsafe.Pointer(c_tag))

	return getError(C.rbd_lock_shared(image.image, c_cookie, c_tag))
}

// Unlock releases a lock on the image.
//
// Implements:
//  int rbd_lock_shared(rbd_image_t image, const char *cookie, const char *tag);
func (image *Image) Unlock(cookie string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_cookie := C.CString(cookie)
	defer C.free(unsafe.Pointer(c_cookie))

	return getError(C.rbd_unlock(image.image, c_cookie))
}

// BreakLock forces the release of a lock held by another client.
//
// Implements:
//  int rbd_break_lock(rbd_image_t image, const char *client, const char *cookie);
func (image *Image) BreakLock(client string, cookie string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_client := C.CString(client)
	c_cookie := C.CString(cookie)
	defer C.free(unsafe.Pointer(c_client))
	defer C.free(unsafe.Pointer(c_cookie))

	return getError(C.rbd_break_lock(image.image, c_client, c_cookie))
}

// ssize_t rbd_read(rbd_image_t image, uint64_t ofs, size_t len, char *buf);
// TODO: int64_t rbd_read_iterate(rbd_image_t image, uint64_t ofs, size_t len,
//              int (*cb)(uint64_t, size_t, const char *, void *), void *arg);
// TODO: int rbd_read_iterate2(rbd_image_t image, uint64_t ofs, uint64_t len,
//               int (*cb)(uint64_t, size_t, const char *, void *), void *arg);
// TODO: int rbd_diff_iterate(rbd_image_t image,
//              const char *fromsnapname,
//              uint64_t ofs, uint64_t len,
//              int (*cb)(uint64_t, size_t, int, void *), void *arg);
func (image *Image) Read(data []byte) (int, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	ret := int(C.rbd_read(
		image.image,
		(C.uint64_t)(image.offset),
		(C.size_t)(len(data)),
		(*C.char)(unsafe.Pointer(&data[0]))))

	if ret < 0 {
		return 0, RBDError(ret)
	}

	image.offset += int64(ret)
	if ret < len(data) {
		return ret, io.EOF
	}

	return ret, nil
}

// ssize_t rbd_write(rbd_image_t image, uint64_t ofs, size_t len, const char *buf);
func (image *Image) Write(data []byte) (n int, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	ret := int(C.rbd_write(image.image, C.uint64_t(image.offset),
		C.size_t(len(data)), (*C.char)(unsafe.Pointer(&data[0]))))

	if ret >= 0 {
		image.offset += int64(ret)
	}

	if ret != len(data) {
		err = RBDError(-C.EPERM)
	}

	return ret, err
}

func (image *Image) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case SeekSet:
		image.offset = offset
	case SeekCur:
		image.offset += offset
	case SeekEnd:
		stats, err := image.Stat()
		if err != nil {
			return 0, err
		}
		image.offset = int64(stats.Size) - offset
	default:
		return 0, errors.New("Wrong value for whence")
	}
	return image.offset, nil
}

// int rbd_discard(rbd_image_t image, uint64_t ofs, uint64_t len);
func (image *Image) Discard(ofs uint64, length uint64) (int, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	ret := C.rbd_discard(image.image, C.uint64_t(ofs), C.uint64_t(length))
	if ret < 0 {
		return 0, RBDError(ret)
	}

	return int(ret), nil
}

func (image *Image) ReadAt(data []byte, off int64) (int, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	ret := int(C.rbd_read(
		image.image,
		(C.uint64_t)(off),
		(C.size_t)(len(data)),
		(*C.char)(unsafe.Pointer(&data[0]))))

	if ret < 0 {
		return 0, RBDError(ret)
	}

	if ret < len(data) {
		return ret, io.EOF
	}

	return ret, nil
}

func (image *Image) WriteAt(data []byte, off int64) (n int, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, nil
	}

	ret := int(C.rbd_write(image.image, C.uint64_t(off),
		C.size_t(len(data)), (*C.char)(unsafe.Pointer(&data[0]))))

	if ret != len(data) {
		err = RBDError(-C.EPERM)
	}

	return ret, err
}

// int rbd_flush(rbd_image_t image);
func (image *Image) Flush() error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	return getError(C.rbd_flush(image.image))
}

// int rbd_snap_list(rbd_image_t image, rbd_snap_info_t *snaps, int *max_snaps);
// void rbd_snap_list_end(rbd_snap_info_t *snaps);
func (image *Image) GetSnapshotNames() (snaps []SnapInfo, err error) {
	if err := image.validate(imageIsOpen); err != nil {
		return nil, err
	}

	var c_max_snaps C.int

	ret := C.rbd_snap_list(image.image, nil, &c_max_snaps)

	c_snaps := make([]C.rbd_snap_info_t, c_max_snaps)
	snaps = make([]SnapInfo, c_max_snaps)

	ret = C.rbd_snap_list(image.image,
		&c_snaps[0], &c_max_snaps)
	if ret < 0 {
		return nil, RBDError(ret)
	}

	for i, s := range c_snaps {
		snaps[i] = SnapInfo{Id: uint64(s.id),
			Size: uint64(s.size),
			Name: C.GoString(s.name)}
	}

	C.rbd_snap_list_end(&c_snaps[0])
	return snaps[:len(snaps)-1], nil
}

// int rbd_snap_create(rbd_image_t image, const char *snapname);
func (image *Image) CreateSnapshot(snapname string) (*Snapshot, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return nil, err
	}

	c_snapname := C.CString(snapname)
	defer C.free(unsafe.Pointer(c_snapname))

	ret := C.rbd_snap_create(image.image, c_snapname)
	if ret < 0 {
		return nil, RBDError(ret)
	}

	return &Snapshot{
		image: image,
		name:  snapname,
	}, nil
}

//
func (image *Image) GetSnapshot(snapname string) *Snapshot {
	return &Snapshot{
		image: image,
		name:  snapname,
	}
}

// int rbd_get_parent_info(rbd_image_t image,
//  char *parent_pool_name, size_t ppool_namelen, char *parent_name,
//  size_t pnamelen, char *parent_snap_name, size_t psnap_namelen)
func (image *Image) GetParentInfo(p_pool, p_name, p_snapname []byte) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	ret := C.rbd_get_parent_info(
		image.image,
		(*C.char)(unsafe.Pointer(&p_pool[0])),
		(C.size_t)(len(p_pool)),
		(*C.char)(unsafe.Pointer(&p_name[0])),
		(C.size_t)(len(p_name)),
		(*C.char)(unsafe.Pointer(&p_snapname[0])),
		(C.size_t)(len(p_snapname)))
	if ret == 0 {
		return nil
	} else {
		return RBDError(ret)
	}
}

// int rbd_metadata_get(rbd_image_t image, const char *key, char *value, size_t *vallen)
func (image *Image) GetMetadata(key string) (string, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return "", err
	}

	c_key := C.CString(key)
	defer C.free(unsafe.Pointer(c_key))

	var c_vallen C.size_t
	ret := C.rbd_metadata_get(image.image, c_key, nil, (*C.size_t)(&c_vallen))
	// get size of value
	// ret -34 because we pass nil as value pointer
	if ret != 0 && ret != -C.ERANGE {
		return "", RBDError(ret)
	}

	// make a bytes array with a good size
	value := make([]byte, c_vallen-1)
	ret = C.rbd_metadata_get(image.image, c_key, (*C.char)(unsafe.Pointer(&value[0])), (*C.size_t)(&c_vallen))
	if ret < 0 {
		return "", RBDError(ret)
	}

	return string(value), nil
}

// int rbd_metadata_set(rbd_image_t image, const char *key, const char *value)
func (image *Image) SetMetadata(key string, value string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_key := C.CString(key)
	c_value := C.CString(value)
	defer C.free(unsafe.Pointer(c_key))
	defer C.free(unsafe.Pointer(c_value))

	ret := C.rbd_metadata_set(image.image, c_key, c_value)
	if ret < 0 {
		return RBDError(ret)
	}

	return nil
}

// int rbd_metadata_remove(rbd_image_t image, const char *key)
func (image *Image) RemoveMetadata(key string) error {
	if err := image.validate(imageIsOpen); err != nil {
		return err
	}

	c_key := C.CString(key)
	defer C.free(unsafe.Pointer(c_key))

	ret := C.rbd_metadata_remove(image.image, c_key)
	if ret < 0 {
		return RBDError(ret)
	}

	return nil
}

// GetId returns the internal image ID string.
//
// Implements:
//  int rbd_get_id(rbd_image_t image, char *id, size_t id_len);
func (image *Image) GetId() (string, error) {
	if err := image.validate(imageIsOpen); err != nil {
		return "", err
	}
	size := C.size_t(1024)
	buf := make([]byte, size)
	for {
		ret := C.rbd_get_id(
			image.image,
			(*C.char)(unsafe.Pointer(&buf[0])),
			size)
		if ret == -C.ERANGE && size <= 8192 {
			size *= 2
			buf = make([]byte, size)
		} else if ret < 0 {
			return "", getError(ret)
		}
		id := C.GoString((*C.char)(unsafe.Pointer(&buf[0])))
		return id, nil
	}
}

// int rbd_snap_remove(rbd_image_t image, const char *snapname);
func (snapshot *Snapshot) Remove() error {
	if err := snapshot.validate(snapshotNeedsName | imageIsOpen); err != nil {
		return err
	}

	c_snapname := C.CString(snapshot.name)
	defer C.free(unsafe.Pointer(c_snapname))

	return getError(C.rbd_snap_remove(snapshot.image.image, c_snapname))
}

// int rbd_snap_rollback(rbd_image_t image, const char *snapname);
// int rbd_snap_rollback_with_progress(rbd_image_t image, const char *snapname,
//                  librbd_progress_fn_t cb, void *cbdata);
func (snapshot *Snapshot) Rollback() error {
	if err := snapshot.validate(snapshotNeedsName | imageIsOpen); err != nil {
		return err
	}

	c_snapname := C.CString(snapshot.name)
	defer C.free(unsafe.Pointer(c_snapname))

	return getError(C.rbd_snap_rollback(snapshot.image.image, c_snapname))
}

// int rbd_snap_protect(rbd_image_t image, const char *snap_name);
func (snapshot *Snapshot) Protect() error {
	if err := snapshot.validate(snapshotNeedsName | imageIsOpen); err != nil {
		return err
	}

	c_snapname := C.CString(snapshot.name)
	defer C.free(unsafe.Pointer(c_snapname))

	return getError(C.rbd_snap_protect(snapshot.image.image, c_snapname))
}

// int rbd_snap_unprotect(rbd_image_t image, const char *snap_name);
func (snapshot *Snapshot) Unprotect() error {
	if err := snapshot.validate(snapshotNeedsName | imageIsOpen); err != nil {
		return err
	}

	c_snapname := C.CString(snapshot.name)
	defer C.free(unsafe.Pointer(c_snapname))

	return getError(C.rbd_snap_unprotect(snapshot.image.image, c_snapname))
}

// int rbd_snap_is_protected(rbd_image_t image, const char *snap_name,
//               int *is_protected);
func (snapshot *Snapshot) IsProtected() (bool, error) {
	if err := snapshot.validate(snapshotNeedsName | imageIsOpen); err != nil {
		return false, err
	}

	var c_is_protected C.int

	c_snapname := C.CString(snapshot.name)
	defer C.free(unsafe.Pointer(c_snapname))

	ret := C.rbd_snap_is_protected(snapshot.image.image, c_snapname,
		&c_is_protected)
	if ret < 0 {
		return false, RBDError(ret)
	}

	return c_is_protected != 0, nil
}

// int rbd_snap_set(rbd_image_t image, const char *snapname);
func (snapshot *Snapshot) Set() error {
	if err := snapshot.validate(snapshotNeedsName | imageIsOpen); err != nil {
		return err
	}

	c_snapname := C.CString(snapshot.name)
	defer C.free(unsafe.Pointer(c_snapname))

	return getError(C.rbd_snap_set(snapshot.image.image, c_snapname))
}

// GetTrashList returns a slice of TrashInfo structs, containing information about all RBD images
// currently residing in the trash.
func GetTrashList(ioctx *rados.IOContext) ([]TrashInfo, error) {
	var num_entries C.size_t

	// Call rbd_trash_list with nil pointer to get number of trash entries.
	if C.rbd_trash_list(C.rados_ioctx_t(ioctx.Pointer()), nil, &num_entries); num_entries == 0 {
		return nil, nil
	}

	c_entries := make([]C.rbd_trash_image_info_t, num_entries)
	trashList := make([]TrashInfo, num_entries)

	if ret := C.rbd_trash_list(C.rados_ioctx_t(ioctx.Pointer()), &c_entries[0], &num_entries); ret < 0 {
		return nil, RBDError(ret)
	}

	for i, ti := range c_entries {
		trashList[i] = TrashInfo{
			Id:               C.GoString(ti.id),
			Name:             C.GoString(ti.name),
			DeletionTime:     time.Unix(int64(ti.deletion_time), 0),
			DefermentEndTime: time.Unix(int64(ti.deferment_end_time), 0),
		}
	}

	// Free rbd_trash_image_info_t pointers
	C.rbd_trash_list_cleanup(&c_entries[0], num_entries)

	return trashList, nil
}

// TrashRemove permanently deletes the trashed RBD with the specified id.
func TrashRemove(ioctx *rados.IOContext, id string, force bool) error {
	c_id := C.CString(id)
	defer C.free(unsafe.Pointer(c_id))

	return getError(C.rbd_trash_remove(C.rados_ioctx_t(ioctx.Pointer()), c_id, C.bool(force)))
}

// TrashRestore restores the trashed RBD with the specified id back to the pool from whence it
// came, with the specified new name.
func TrashRestore(ioctx *rados.IOContext, id, name string) error {
	c_id := C.CString(id)
	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_id))
	defer C.free(unsafe.Pointer(c_name))

	return getError(C.rbd_trash_restore(C.rados_ioctx_t(ioctx.Pointer()), c_id, c_name))
}

// OpenImage will open an existing rbd image by name and snapshot name,
// returning a new opened image. Pass the NoSnapshot sentinel value as the
// snapName to explicitly indicate that no snapshot name is being provided.
//
// Implements:
//  int rbd_open(rados_ioctx_t io, const char *name,
//               rbd_image_t *image, const char *snap_name);
func OpenImage(ioctx *rados.IOContext, name, snapName string) (*Image, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cSnapName *C.char
	if snapName != NoSnapshot {
		cSnapName = C.CString(snapName)
		defer C.free(unsafe.Pointer(cSnapName))
	}

	var cImage C.rbd_image_t
	ret := C.rbd_open(
		C.rados_ioctx_t(ioctx.Pointer()),
		cName,
		&cImage,
		cSnapName)

	if ret != 0 {
		return nil, getError(ret)
	}

	return &Image{
		ioctx: ioctx,
		name:  name,
		image: cImage,
	}, nil
}

// OpenImageReadOnly will open an existing rbd image by name and snapshot name,
// returning a new opened-for-read image.  Pass the NoSnapshot sentinel value
// as the snapName to explicitly indicate that no snapshot name is being
// provided.
//
// Implements:
//  int rbd_open_read_only(rados_ioctx_t io, const char *name,
//                         rbd_image_t *image, const char *snap_name);
func OpenImageReadOnly(ioctx *rados.IOContext, name, snapName string) (*Image, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var cSnapName *C.char
	if snapName != NoSnapshot {
		cSnapName = C.CString(snapName)
		defer C.free(unsafe.Pointer(cSnapName))
	}

	var cImage C.rbd_image_t
	ret := C.rbd_open_read_only(
		C.rados_ioctx_t(ioctx.Pointer()),
		cName,
		&cImage,
		cSnapName)

	if ret != 0 {
		return nil, getError(ret)
	}

	return &Image{
		ioctx: ioctx,
		name:  name,
		image: cImage,
	}, nil
}

// OpenImageById will open an existing rbd image by ID and snapshot name,
// returning a new opened image. Pass the NoSnapshot sentinel value as the
// snapName to explicitly indicate that no snapshot name is being provided.
// Error handling will fail & segfault unless compiled with a version of ceph
// that fixes https://tracker.ceph.com/issues/43178
//
// Implements:
//  int rbd_open_by_id(rados_ioctx_t io, const char *id,
//                     rbd_image_t *image, const char *snap_name);
func OpenImageById(ioctx *rados.IOContext, id, snapName string) (*Image, error) {
	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))

	var cSnapName *C.char
	if snapName != NoSnapshot {
		cSnapName = C.CString(snapName)
		defer C.free(unsafe.Pointer(cSnapName))
	}

	var cImage C.rbd_image_t
	ret := C.rbd_open_by_id(
		C.rados_ioctx_t(ioctx.Pointer()),
		cid,
		&cImage,
		cSnapName)

	if ret != 0 {
		return nil, getError(ret)
	}

	return &Image{
		ioctx: ioctx,
		image: cImage,
	}, nil
}

// OpenImageByIdReadOnly will open an existing rbd image by ID and snapshot
// name, returning a new opened-for-read image. Pass the NoSnapshot sentinel
// value as the snapName to explicitly indicate that no snapshot name is being
// provided.
// Error handling will fail & segfault unless compiled with a version of ceph
// that fixes https://tracker.ceph.com/issues/43178
//
// Implements:
//  int rbd_open_by_id_read_only(rados_ioctx_t io, const char *id,
//                               rbd_image_t *image, const char *snap_name);
func OpenImageByIdReadOnly(ioctx *rados.IOContext, id, snapName string) (*Image, error) {
	cid := C.CString(id)
	defer C.free(unsafe.Pointer(cid))

	var cSnapName *C.char
	if snapName != NoSnapshot {
		cSnapName = C.CString(snapName)
		defer C.free(unsafe.Pointer(cSnapName))
	}

	var cImage C.rbd_image_t
	ret := C.rbd_open_by_id_read_only(
		C.rados_ioctx_t(ioctx.Pointer()),
		cid,
		&cImage,
		cSnapName)

	if ret != 0 {
		return nil, getError(ret)
	}

	return &Image{
		ioctx: ioctx,
		image: cImage,
	}, nil
}

// CreateImage creates a new rbd image using provided image options.
//
// Implements:
//  int rbd_create4(rados_ioctx_t io, const char *name, uint64_t size,
//                 rbd_image_options_t opts);
func CreateImage(ioctx *rados.IOContext, name string, size uint64, rio *RbdImageOptions) error {

	if rio == nil {
		return RBDError(C.EINVAL)
	}

	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_name))

	ret := C.rbd_create4(C.rados_ioctx_t(ioctx.Pointer()), c_name,
		C.uint64_t(size), C.rbd_image_options_t(rio.options))
	return getError(ret)
}

// RemoveImage removes the specified rbd image.
//
// Implements:
//  int rbd_remove(rados_ioctx_t io, const char *name);
func RemoveImage(ioctx *rados.IOContext, name string) error {
	c_name := C.CString(name)
	defer C.free(unsafe.Pointer(c_name))
	return getError(C.rbd_remove(C.rados_ioctx_t(ioctx.Pointer()), c_name))
}
