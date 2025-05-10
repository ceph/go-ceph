//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsersAndGroupsIdentity(t *testing.T) {
	ug := NewUsersAndGroups("ug1")
	assert.Equal(t, ug.Type(), UsersAndGroupsType)
	ugid := ug.Identity()
	assert.Equal(t, ugid.Type(), UsersAndGroupsType)
	assert.Equal(t, ugid.String(), "ceph.smb.usersgroups.ug1")
	assert.Equal(t, ug.Intent(), Present)
}

func TestUsersAndGroupsValidate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		ug := &UsersAndGroups{}
		assert.ErrorContains(t, ug.Validate(), "intent")
	})
	t.Run("missingID", func(t *testing.T) {
		ug := &UsersAndGroups{IntentValue: Present}
		assert.ErrorContains(t, ug.Validate(), "UsersGroupsID")
	})
	t.Run("intentRemoved", func(t *testing.T) {
		ug := &UsersAndGroups{IntentValue: Removed, UsersGroupsID: "boop"}
		assert.NoError(t, ug.Validate())
	})

	// from this point on we're going to progressively add to
	// ug1 until it is valid
	ug1 := &UsersAndGroups{
		IntentValue:   Present,
		UsersGroupsID: "hugh1",
	}
	t.Run("missingUsersAndGroupsValues", func(t *testing.T) {
		assert.ErrorContains(t, ug1.Validate(), "Values")
	})

	ug1.Values = &UsersAndGroupsValues{}
	t.Run("missingUsersSlice", func(t *testing.T) {
		assert.ErrorContains(t, ug1.Validate(), "Users")
	})

	ug1.Values.Users = []UserInfo{
		{Name: "alice", Password: "l3tm31n"},
		{Name: "bob", Password: "1122334455"},
	}
	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, ug1.Validate())
	})
}

func TestUsersAndGroupsNewSetVals(t *testing.T) {
	ug := NewUsersAndGroups("hug1").SetValues(
		[]UserInfo{
			{Name: "alice", Password: "l3tm31n"},
			{Name: "bob", Password: "1122334455"},
		},
		[]GroupInfo{
			{Name: "itstaff"},
		},
	)
	assert.Equal(t, ug.Intent(), Present)
	assert.Equal(t, ug.UsersGroupsID, "hug1")
	assert.Len(t, ug.Values.Users, 2)
	assert.Len(t, ug.Values.Groups, 1)
	assert.Equal(t, ug.LinkedToCluster, "")
}

func TestUsersAndGroupsNewToRemove(t *testing.T) {
	rc := NewUsersAndGroupsToRemove("nope")
	assert.Equal(t, rc.Intent(), Removed)
	assert.Equal(t, rc.UsersGroupsID, "nope")
}

func TestUsersAndGroupsNewLinked(t *testing.T) {
	c := NewUserCluster("c1")
	ug := NewLinkedUsersAndGroups(c).SetValues(
		[]UserInfo{
			{Name: "alice", Password: "l3tm31n"},
			{Name: "bob", Password: "1122334455"},
		},
		[]GroupInfo{
			{Name: "itstaff"},
		},
	)
	assert.NoError(t, ug.Validate())
	assert.Equal(t, ug.LinkedToCluster, "c1")
	assert.Len(t, c.UserGroupSettings, 1)
	assert.Equal(t, c.UserGroupSettings[0].Ref, ug.UsersGroupsID)
}

func TestUsersAndGroupsMarshalUnmarshal(t *testing.T) {
	ug := NewUsersAndGroups("ug1").SetValues(
		[]UserInfo{
			{Name: "alice", Password: "l3tm31n"},
			{Name: "bob", Password: "1122334455"},
		},
		[]GroupInfo{
			{Name: "itstaff"},
		},
	)
	j, err := json.Marshal(ug)
	assert.NoError(t, err)
	ug2 := &UsersAndGroups{}
	err = json.Unmarshal(j, ug2)
	assert.NoError(t, err)
	assert.Equal(t, ug.UsersGroupsID, ug2.UsersGroupsID)
	assert.EqualValues(t, ug.Values, ug2.Values)
}

func TestUsersAndGroupsSetAuth(t *testing.T) {
	ug := &UsersAndGroups{IntentValue: Present, UsersGroupsID: "abc"}
	ug.SetValues(
		[]UserInfo{
			{Name: "alice", Password: "l3tm31n"},
			{Name: "bob", Password: "1122334455"},
		},
		[]GroupInfo{
			{Name: "itstaff"},
		},
	)
	assert.NotNil(t, ug.Values)
	assert.Len(t, ug.Values.Users, 2)
	assert.Len(t, ug.Values.Groups, 1)
}
