//go:build !nautilus && !ceph_preview
// +build !nautilus,!ceph_preview

package rbd

import (
	"testing"
)

// testDescriptionReplayStatus is a stub function that exists only to be
// compiled as a near no-op on non ceph_preview builds.
func testDescriptionReplayStatus(t *testing.T, _ SiteMirrorImageStatus) {
	t.Log("not testing DescriptionReplayStatus")
}
