// +build !luminous

package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFeaturesInMimic(t *testing.T) {
	f, ok := featureNameToBit[FeatureNameOperations]
	assert.True(t, ok)
	assert.Equal(t, f, FeatureOperations)
}
