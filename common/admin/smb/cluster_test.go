//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClusterIdentity(t *testing.T) {
	cluster := NewUserCluster("c1")
	assert.Equal(t, cluster.Type(), ClusterType)
	cid := cluster.Identity()
	assert.Equal(t, cid.Type(), ClusterType)
	assert.Equal(t, cid.String(), "ceph.smb.cluster.c1")
	assert.Equal(t, cluster.Intent(), Present)
}

func TestClusterValidate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		c := &Cluster{}
		assert.ErrorContains(t, c.Validate(), "intent")
	})
	t.Run("missingID", func(t *testing.T) {
		c := &Cluster{IntentValue: Present}
		assert.ErrorContains(t, c.Validate(), "ClusterID")
	})
	t.Run("intentRemoved", func(t *testing.T) {
		c := &Cluster{IntentValue: Removed, ClusterID: "fred"}
		assert.NoError(t, c.Validate())
	})

	// from this point on we're going to progressively add to
	// c1 until it is valid
	c1 := &Cluster{
		IntentValue: Present,
		ClusterID:   "bob",
	}
	t.Run("missingAuthMode", func(t *testing.T) {
		assert.ErrorContains(t, c1.Validate(), "AuthMode")
	})
	c1.AuthMode = ActiveDirectoryAuth
	t.Run("missingDomainSettings", func(t *testing.T) {
		assert.ErrorContains(t, c1.Validate(), "domain")
	})
	c1.AuthMode = UserAuth
	t.Run("missingUserGroupSettings", func(t *testing.T) {
		assert.ErrorContains(t, c1.Validate(), "user")
	})

	c1.UserGroupSettings = []UserGroupSource{{ResourceSource, "foo"}}
	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, c1.Validate())
	})
}

func TestClusterNewActiveDirectory(t *testing.T) {
	// ensure that the new method for ad clusters creates
	// the correct defaults
	cluster := NewActiveDirectoryCluster("davey", "cool.example.com", "j1", "j2")
	assert.Equal(t, cluster.IntentValue, Present)
	assert.Equal(t, cluster.ClusterID, "davey")
	assert.Equal(t, cluster.AuthMode, ActiveDirectoryAuth)
	assert.Equal(t, cluster.DomainSettings.Realm, "cool.example.com")
	assert.Len(t, cluster.DomainSettings.JoinSources, 2)
	assert.Equal(t, cluster.DomainSettings.JoinSources[0].Ref, "j1")
	assert.Equal(t, cluster.DomainSettings.JoinSources[1].Ref, "j2")
}

func TestClusterNewToRemove(t *testing.T) {
	rc := NewClusterToRemove("nope")
	assert.Equal(t, rc.Intent(), Removed)
	assert.Equal(t, rc.ClusterID, "nope")
}

func TestMarshalCluster(t *testing.T) {
	cluster := NewUserCluster("c1")
	_, err := json.Marshal(cluster)
	assert.NoError(t, err)
}

func TestMarshalUnmarshalCluster(t *testing.T) {
	c := NewUserCluster("joey", "pibb")
	assert.NoError(t, c.Validate())
	j, err := json.Marshal(c)
	assert.NoError(t, err)
	c2 := &Cluster{}
	err = json.Unmarshal(j, c2)
	assert.NoError(t, err)
	assert.NoError(t, c2.Validate())
	assert.EqualValues(t, c.ClusterID, c2.ClusterID)
	assert.EqualValues(t, c.IntentValue, c2.IntentValue)
	assert.EqualValues(t, c.UserGroupSettings, c2.UserGroupSettings)
}

func TestMarshalUnmarshalClusterLinkedUsers(t *testing.T) {
	c := NewUserCluster("ceeone")
	ug := NewLinkedUsersAndGroups(c).SetValues(
		[]UserInfo{
			{"bob", "M4r13y"},
			{"billy", "p14n0m4N"},
		},
		[]GroupInfo{{"clients"}},
	)
	rg := resourceGroup{Resources: []Resource{c, ug}}
	assert.NoError(t, ValidateResources(rg.Resources))
	j, err := json.Marshal(rg)
	assert.NoError(t, err)

	rg = resourceGroup{}
	err = json.Unmarshal(j, &rg)
	assert.NoError(t, err)
	assert.Len(t, rg.Resources, 2)
	assert.Equal(t, rg.Resources[0].Type(), ClusterType)
	assert.Equal(t, rg.Resources[1].Type(), UsersAndGroupsType)
}

func TestClusterSimplePlacement(t *testing.T) {
	t.Run("countOnly", func(t *testing.T) {
		p := SimplePlacement(3, "") // count of 3, no label
		c := NewUserCluster("joey", "pibb").SetPlacement(p)
		assert.Len(t, c.Placement, 1)
		assert.EqualValues(t, c.Placement["count"], 3)
	})
	t.Run("labelOnly", func(t *testing.T) {
		p := SimplePlacement(0, "ilovesmb") // count of 3, no label
		c := NewUserCluster("joey", "pibb").SetPlacement(p)
		assert.Len(t, c.Placement, 1)
		assert.EqualValues(t, c.Placement["label"], "ilovesmb")
	})
	t.Run("marshalCountAndLabel", func(t *testing.T) {
		p := SimplePlacement(3, "ilovesmb") // count of 3, no label
		c := NewUserCluster("joey", "pibb").SetPlacement(p)
		j, err := json.Marshal(c)
		assert.NoError(t, err)

		c2 := &Cluster{}
		err = json.Unmarshal(j, c2)
		assert.NoError(t, err)
		assert.Len(t, c2.Placement, 2)
		assert.EqualValues(t, c.Placement["count"], 3)
		assert.EqualValues(t, c.Placement["label"], "ilovesmb")
	})
}
