//go:build ceph_preview

package rados

import (
	"sort"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestGetOmapValuesOrdered() {
	suite.SetupConnection()

	orig := map[string][]byte{
		"key1":          []byte("value1"),
		"key2":          []byte("value2"),
		"key3":          []byte("value3"),
		"prefixed-key4": []byte("value4"),
		"prefixed-key5": []byte("value5"),
	}

	oid := suite.GenObjectName()
	err := suite.ioctx.SetOmap(oid, orig)
	assert.NoError(suite.T(), err)

	// full ordered fetch
	fetched, err := suite.ioctx.GetOmapValuesOrdered(oid, "", "", 100)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), fetched, len(orig))

	// results must be sorted by key and match the original values
	assert.True(suite.T(), sort.SliceIsSorted(fetched, func(i, j int) bool {
		return fetched[i].Key < fetched[j].Key
	}), "returned slice must be ordered by key")
	for _, kv := range fetched {
		assert.Equal(suite.T(), orig[kv.Key], kv.Value)
	}

	// filter by prefix
	fetched, err = suite.ioctx.GetOmapValuesOrdered(oid, "", "prefixed", 100)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []OmapKeyValue{
		{Key: "prefixed-key4", Value: []byte("value4")},
		{Key: "prefixed-key5", Value: []byte("value5")},
	}, fetched)

	// maxReturn caps the number of returned pairs, in order
	fetched, err = suite.ioctx.GetOmapValuesOrdered(oid, "", "", 2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []OmapKeyValue{
		{Key: "key1", Value: []byte("value1")},
		{Key: "key2", Value: []byte("value2")},
	}, fetched)

	// startAfter based pagination walks the omap in order without gaps or
	// overlaps, using the last returned key as the next startAfter.
	var paged []OmapKeyValue
	startAfter := ""
	for {
		page, err := suite.ioctx.GetOmapValuesOrdered(oid, startAfter, "", 2)
		assert.NoError(suite.T(), err)
		if len(page) == 0 {
			break
		}
		paged = append(paged, page...)
		startAfter = page[len(page)-1].Key
	}
	assert.Equal(suite.T(), fetchAllExpected(orig), paged)
}

func (suite *RadosTestSuite) TestGetAllOmapValuesOrdered() {
	suite.SetupConnection()

	orig := map[string][]byte{
		"key1":          []byte("value1"),
		"key2":          []byte("value2"),
		"key3":          []byte("value3"),
		"prefixed-key4": []byte("value4"),
		"prefixed-key5": []byte("value5"),
	}

	oid := suite.GenObjectName()
	err := suite.ioctx.SetOmap(oid, orig)
	assert.NoError(suite.T(), err)

	expected := fetchAllExpected(orig)

	// iterator size bigger than the map size
	fetched, err := suite.ioctx.GetAllOmapValuesOrdered(oid, "", "", 100)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, fetched)

	// iterator size smaller than the map size (forces multiple iterations)
	fetched, err = suite.ioctx.GetAllOmapValuesOrdered(oid, "", "", 2)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expected, fetched)

	// empty omap returns an empty (non-nil) slice
	err = suite.ioctx.CleanOmap(oid)
	assert.NoError(suite.T(), err)

	fetched, err = suite.ioctx.GetAllOmapValuesOrdered(oid, "", "", 100)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), []OmapKeyValue{}, fetched)
}

// fetchAllExpected returns the given map as a slice of OmapKeyValue sorted by
// key, matching the ordering that RADOS returns.
func fetchAllExpected(m map[string][]byte) []OmapKeyValue {
	out := make([]OmapKeyValue, 0, len(m))
	for k, v := range m {
		out = append(out, OmapKeyValue{Key: k, Value: v})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Key < out[j].Key
	})
	return out
}
