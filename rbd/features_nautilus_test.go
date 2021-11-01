package rbd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetFeaturesInNautilus(t *testing.T) {
	f, ok := featureNameToBit[FeatureNameMigrating]
	assert.True(t, ok)
	assert.Equal(t, f, FeatureMigrating)
}
