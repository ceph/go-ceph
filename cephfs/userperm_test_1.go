package cephfs

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserPerm(t *testing.T) {
	t.Run("newAndDestroy", func(t *testing.T) {
		uperm := NewUserPerm(0, 0, nil)
		assert.NotNil(t, uperm)
		assert.Equal(t, 0, len(uperm.gidList))
		assert.True(t, uperm.managed)

		// Destroy should be idempotent in our go library
		uperm.Destroy()
		uperm.Destroy()
	})
	t.Run("notManagedDestroy", func(t *testing.T) {
		uperm := &UserPerm{}
		assert.False(t, uperm.managed)
		// Calling destroy shouldn't do much but is safe to call (many times)
		uperm.Destroy()
		uperm.Destroy()
	})
	t.Run("tryForceGc", func(_ *testing.T) {
		func() {
			uperm := NewUserPerm(0, 0, nil)
			_ = uperm
		}()
		runtime.GC()
	})
	t.Run("gidList", func(t *testing.T) {
		uperm := NewUserPerm(1000, 1000, []int{1028, 1192, 2112})
		defer uperm.Destroy()
		assert.Equal(t, 3, len(uperm.gidList))
		assert.EqualValues(t, 1028, uperm.gidList[0])
		assert.EqualValues(t, 1192, uperm.gidList[1])
		assert.EqualValues(t, 2112, uperm.gidList[2])
	})
}
