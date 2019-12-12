package rbd_test

import (
	"github.com/ceph/go-ceph/rbd"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRbdOptions(t *testing.T) {
	var i uint64
	var s string
	var err error

	options := rbd.NewRbdImageOptions()
	defer options.Destroy()

	err = options.SetUint64(rbd.RbdImageOptionFormat, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionFormat, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionFormat)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionFormat)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionFeatures, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionFeatures, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionFeatures)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionFeatures)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionOrder, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionOrder, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionOrder)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionOrder)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionStripeUnit, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionStripeUnit, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionStripeUnit)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionStripeUnit)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionStripeCount, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionStripeCount, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionStripeCount)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionStripeCount)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionJournalOrder, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionJournalOrder, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionJournalOrder)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionJournalOrder)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionJournalSplayWidth, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionJournalSplayWidth, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionJournalSplayWidth)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionJournalSplayWidth)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionJournalPool, 1)
	assert.Error(t, err)
	err = options.SetString(rbd.RbdImageOptionJournalPool, "journal")
	assert.NoError(t, err)
	_, err = options.GetUint64(rbd.RbdImageOptionJournalPool)
	assert.Error(t, err)
	s, err = options.GetString(rbd.RbdImageOptionJournalPool)
	assert.NoError(t, err)
	assert.True(t, s == "journal")

	err = options.SetUint64(rbd.RbdImageOptionFeaturesSet, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionFeaturesSet, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionFeaturesSet)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionFeaturesSet)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionFeaturesClear, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionFeaturesClear, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionFeaturesClear)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionFeaturesClear)
	assert.Error(t, err)

	err = options.SetUint64(rbd.RbdImageOptionDataPool, 1)
	assert.Error(t, err)
	err = options.SetString(rbd.RbdImageOptionDataPool, "data")
	assert.NoError(t, err)
	_, err = options.GetUint64(rbd.RbdImageOptionDataPool)
	assert.Error(t, err)
	s, err = options.GetString(rbd.RbdImageOptionDataPool)
	assert.NoError(t, err)
	assert.True(t, s == "data")

	/* introduced with Ceph Mimic, can not be tested on Luminous
	err = options.SetUint64(rbd.RbdImageOptionFlatten, 1)
	assert.NoError(t, err)
	err = options.SetString(rbd.RbdImageOptionFlatten, "string not allowed")
	assert.Error(t, err)
	i, err = options.GetUint64(rbd.RbdImageOptionFlatten)
	assert.NoError(t, err)
	assert.True(t, i == 1)
	_, err = options.GetString(rbd.RbdImageOptionFlatten)
	assert.Error(t, err)
	*/
}

func TestRbdOptionsClear(t *testing.T) {
	options := rbd.NewRbdImageOptions()

	// set at least one option
	err := options.SetUint64(rbd.RbdImageOptionFormat, 1)
	assert.NoError(t, err)

	empty := options.IsEmpty()
	assert.False(t, empty)

	options.Clear()
	empty = options.IsEmpty()
	assert.True(t, empty)

	options.Destroy()
}
