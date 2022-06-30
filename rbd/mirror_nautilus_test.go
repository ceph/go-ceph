//go:build nautilus
// +build nautilus

package rbd

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ceph/go-ceph/rados"
)

const (
	mirrorPeerCluster    = "cluster2"
	mirrorPeerClientName = "client.rbd-mirror-remote01"
)

func mustCreateAndOpenImage(t *testing.T, ioctx *rados.IOContext) (*Image, string) {
	opts := NewRbdImageOptions()
	defer opts.Destroy()

	err := opts.SetUint64(ImageOptionFormat, 2)
	require.NoError(t, err)

	name := GetUUID()
	err = CreateImage(ioctx, name, 1<<25, opts)
	require.NoError(t, err)

	img, err := OpenImage(ioctx, name, "")
	require.NoError(t, err)

	err = img.UpdateFeatures(FeatureJournaling, true)
	require.NoError(t, err)

	return img, name
}

func TestMirrorPoolOps(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer conn.DeletePool(poolName)

	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)

	var peerUUID string
	// Note: These test cases can't run in parallel, since each depends on the
	// state enacted by the previous.
	testCases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "get mirror mode with mirroring disabled",
			fn:   testMirrorModeGet(ioctx, MirrorModeDisabled),
		},
		{
			name: "enable mirroring in image mode",
			fn: func(t *testing.T) {
				err := MirrorModeSet(ioctx, MirrorModeImage)
				require.NoError(t, err)
			},
		},
		{
			name: "get mirror mode in image mode",
			fn:   testMirrorModeGet(ioctx, MirrorModeImage),
		},
		{
			name: "change mirror mode to pool mode",
			fn: func(t *testing.T) {
				err := MirrorModeSet(ioctx, MirrorModePool)
				require.NoError(t, err)
			},
		},
		{
			name: "get mirror mode in pool mode",
			fn:   testMirrorModeGet(ioctx, MirrorModePool),
		},
		{
			name: "list mirror peers with no peers",
			fn: func(t *testing.T) {
				peers, err := MirrorPeerList(ioctx)
				require.NoError(t, err)
				require.Len(t, peers, 0)
			},
		},
		{
			name: "add mirror peer",
			fn: func(t *testing.T) {
				peerUUID, err = MirrorPeerAdd(ioctx, mirrorPeerCluster, mirrorPeerClientName)
				require.NoError(t, err)
			},
		},
		{
			name: "list mirror peers with peer",
			fn:   testMirrorPeerList(ioctx, &peerUUID, mirrorPeerCluster, mirrorPeerClientName),
		},
		{
			name: "remove mirror peer",
			fn: func(t *testing.T) {
				err := MirrorPeerRemove(ioctx, peerUUID)
				require.NoError(t, err)
			},
		},
		{
			name: "list mirror peers after removing peer",
			fn: func(t *testing.T) {
				peers, err := MirrorPeerList(ioctx)
				require.NoError(t, err)
				require.Len(t, peers, 0)
			},
		},
		{
			name: "disable mirroring",
			fn: func(t *testing.T) {
				err := MirrorModeSet(ioctx, MirrorModeDisabled)
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		succeeded := t.Run(tc.name, tc.fn)
		// each test depends on the previous one, so abort if any fails
		if !succeeded {
			t.FailNow()
		}
	}
}

func testMirrorModeGet(ioctx *rados.IOContext, expectedMode MirrorMode) func(*testing.T) {
	return func(t *testing.T) {
		mode, err := MirrorModeGet(ioctx)
		require.NoError(t, err)
		require.Equal(t, expectedMode, mode)
	}
}

func testMirrorPeerList(ioctx *rados.IOContext, expectedUUID *string, expectedCluster, expectedClientName string) func(*testing.T) {
	return func(t *testing.T) {
		peers, err := MirrorPeerList(ioctx)
		require.NoError(t, err)
		require.Len(t, peers, 1)

		// expectedUUID is a pointer because other tests will set it after this
		// closure is returned.
		require.Equal(t, *expectedUUID, peers[0].UUID)
		require.Equal(t, expectedCluster, peers[0].ClusterName)
		require.Equal(t, expectedClientName, peers[0].ClientName)
	}
}

func TestMirrorImageOps(t *testing.T) {
	conn := radosConnect(t)
	defer conn.Shutdown()

	poolName := GetUUID()
	err := conn.MakePool(poolName)
	require.NoError(t, err)
	defer conn.DeletePool(poolName)

	ioctx, err := conn.OpenIOContext(poolName)
	require.NoError(t, err)

	err = MirrorModeSet(ioctx, MirrorModeImage)
	require.NoError(t, err)
	defer func() {
		MirrorModeSet(ioctx, MirrorModeDisabled)
	}()

	image, imageName := mustCreateAndOpenImage(t, ioctx)
	defer func() {
		image.Close()
		image.Remove()
	}()

	testCases := []struct {
		name string
		fn   func(*testing.T)
	}{
		{
			name: "get status with mirroring never enabled",
			fn: testMirrorImageGetStatus(image, imageName, MirrorImageDisabled,
				MirrorImageStatusStateUnknown, false, false),
		},
		{
			name: "list statuses with mirroring never enabled",
			fn:   testMirrorImageList(ioctx, 0),
		},
		{
			name: "enable mirroring",
			fn: func(t *testing.T) {
				err := image.MirrorEnable()
				require.NoError(t, err)
			},
		},
		{
			name: "get status with mirroring enabled",
			fn: testMirrorImageGetStatus(image, imageName, MirrorImageEnabled,
				MirrorImageStatusStateUnknown, true, false),
		},
		{
			name: "list statuses with mirroring enabled",
			fn:   testMirrorImageList(ioctx, 1),
		},
		{
			name: "demote image",
			fn: func(t *testing.T) {
				err := image.MirrorDemote()
				require.NoError(t, err)
			},
		},
		{
			name: "get status after demotion",
			fn: testMirrorImageGetStatus(image, imageName, MirrorImageEnabled,
				MirrorImageStatusStateUnknown, false, false),
		},
		{
			name: "promote image",
			fn: func(t *testing.T) {
				err := image.MirrorPromote(false)
				require.NoError(t, err)
			},
		},
		{
			name: "get status after promotion",
			fn: testMirrorImageGetStatus(image, imageName, MirrorImageEnabled,
				MirrorImageStatusStateUnknown, true, false),
		},
		{
			name: "disable mirroring",
			fn: func(t *testing.T) {
				err := image.MirrorDisable(false)
				require.NoError(t, err)
			},
		},
		{
			name: "get status with mirroring disabled after enabled",
			fn: testMirrorImageGetStatus(image, imageName, MirrorImageDisabled,
				MirrorImageStatusStateUnknown, false, false),
		},
		{
			name: "list statuses with mirroring disabled after enabled",
			fn:   testMirrorImageList(ioctx, 0),
		},
	}

	for _, tc := range testCases {
		succeeded := t.Run(tc.name, tc.fn)
		// each test depends on the previous one, so abort if any fails
		if !succeeded {
			t.FailNow()
		}
	}
}

func testMirrorImageGetStatus(
	image *Image,
	expectedName string,
	expectedState MirrorImageState,
	expectedStatusState MirrorImageStatusState,
	expectedPrimary, expectedUp bool) func(*testing.T) {
	return func(t *testing.T) {
		info, err := image.MirrorGetImage()
		require.NoError(t, err)
		require.Equal(t, expectedName, info.Name)
		require.Equal(t, expectedState, info.State)
		require.Equal(t, expectedStatusState, info.StatusState)
		require.Equal(t, expectedPrimary, info.IsPrimary)
		require.Equal(t, expectedUp, info.IsUp)
	}
}

func testMirrorImageList(ioctx *rados.IOContext, expectedImageCount int) func(t *testing.T) {
	return func(t *testing.T) {
		list, err := MirrorImageList(ioctx)
		require.NoError(t, err)
		require.Len(t, list, expectedImageCount)
	}
}

func TestMirroring(t *testing.T) {
	mconfig := mirrorConfig()
	if mconfig == "" {
		t.Skip("no mirror config env var set")
	}

	// this test assumes the rbd pool already exists and is mirrored
	// this must be set up previously by the CI or manually
	poolName := "rbd"

	connA := radosConnect(t)
	defer connA.Shutdown()
	ioctxA, err := connA.OpenIOContext(poolName)
	require.NoError(t, err)

	connB := radosConnectConfig(t, mconfig)
	defer connB.Shutdown()
	ioctxB, err := connB.OpenIOContext(poolName)
	require.NoError(t, err)

	imageA, imageName := mustCreateAndOpenImage(t, ioctxA)
	defer func() {
		imageA.Close()
		imageA.Remove()
	}()

	err = imageA.MirrorEnable()
	require.NoError(t, err)

	_, err = imageA.Write([]byte("hello world!"))
	require.NoError(t, err)

	var imageB *Image
	for i := 0; i < 10; i++ {
		imageB, err = OpenImage(ioctxB, imageName, NoSnapshot)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err)
	defer func() {
		imageB.Close()
		imageB.Remove()
	}()

	mustWaitForMirrorImageInfo(t, imageA, imageName, MirrorImageEnabled,
		MirrorImageStatusStateStopped, true, true)

	mustWaitForMirrorImageInfo(t, imageB, imageName, MirrorImageEnabled,
		MirrorImageStatusStateReplaying, false, true)

	err = imageA.MirrorDemote()
	require.NoError(t, err)

	mustWaitForMirrorImageInfo(t, imageB, imageName, MirrorImageEnabled,
		MirrorImageStatusStateUnknown, false, true)

	err = imageB.MirrorPromote(false)
	require.NoError(t, err)

	buf := make([]byte, 12)
	_, err = imageB.Read(buf)
	require.NoError(t, err)
	require.Equal(t, "hello world!", string(buf))
}

func mirrorConfig() string {
	return os.Getenv("MIRROR_CONF")
}

func mustWaitForMirrorImageInfo(
	t *testing.T,
	image *Image,
	expectedName string,
	expectedState MirrorImageState,
	expectedStatusState MirrorImageStatusState,
	expectedPrimary, expectedUp bool) {

	var lastErr error
	for i := 0; i < 60; i++ {
		if i > 0 {
			time.Sleep(time.Second)
		}

		lastErr = nil
		info, err := image.MirrorGetImage()
		switch {
		case err != nil:
			lastErr = fmt.Errorf("unexpected error while getting image mirroring info: %s", err)
		case info.Name != expectedName:
			lastErr = fmt.Errorf("expected image name %q got %q", expectedName, info.Name)
		case info.State != expectedState:
			lastErr = fmt.Errorf("expected image state %d got %d", expectedState, info.State)
		case info.StatusState != expectedStatusState:
			lastErr = fmt.Errorf("expected image status state %d got %d", expectedStatusState, info.StatusState)
		case info.IsPrimary != expectedPrimary:
			lastErr = fmt.Errorf("expected image primary %v got %v", expectedPrimary, info.IsPrimary)
		case info.IsUp != expectedUp:
			lastErr = fmt.Errorf("expected image up %v got %v", expectedUp, info.IsUp)
		}

		if lastErr == nil {
			break
		}
	}

	require.NoError(t, lastErr)
}
