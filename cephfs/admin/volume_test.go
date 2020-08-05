// +build !luminous,!mimic

package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListVolumes(t *testing.T) {
	fsa := getFSAdmin(t)

	vl, err := fsa.ListVolumes()
	assert.NoError(t, err)
	assert.Len(t, vl, 1)
	assert.Equal(t, "cephfs", vl[0])
}
