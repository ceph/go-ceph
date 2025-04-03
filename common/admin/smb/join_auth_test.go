//go:build !(octopus || pacific || quincy || reef || squid) && ceph_preview

package smb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinAuthIdentity(t *testing.T) {
	ja := NewJoinAuth("ja1")
	assert.Equal(t, ja.Type(), JoinAuthType)
	jid := ja.Identity()
	assert.Equal(t, jid.Type(), JoinAuthType)
	assert.Equal(t, jid.String(), "ceph.smb.join.auth.ja1")
	assert.Equal(t, ja.Intent(), Present)
}

func TestJoinAuthValidate(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		ja := &JoinAuth{}
		assert.ErrorContains(t, ja.Validate(), "intent")
	})
	t.Run("missingID", func(t *testing.T) {
		ja := &JoinAuth{IntentValue: Present}
		assert.ErrorContains(t, ja.Validate(), "AuthID")
	})
	t.Run("intentRemoved", func(t *testing.T) {
		ja := &JoinAuth{IntentValue: Removed, AuthID: "fred"}
		assert.NoError(t, ja.Validate())
	})

	// from this point on we're going to progressively add to
	// ja1 until it is valid
	ja1 := &JoinAuth{
		IntentValue: Present,
		AuthID:      "jalopy",
	}
	t.Run("missingAuth", func(t *testing.T) {
		assert.ErrorContains(t, ja1.Validate(), "Auth")
	})

	ja1.Auth = &JoinAuthValues{}
	t.Run("missingUsername", func(t *testing.T) {
		assert.ErrorContains(t, ja1.Validate(), "Username")
	})

	ja1.Auth.Username = "admin"
	t.Run("missingPassword", func(t *testing.T) {
		assert.ErrorContains(t, ja1.Validate(), "Password")
	})

	ja1.Auth.Password = "Passw0rd"
	t.Run("ok", func(t *testing.T) {
		assert.NoError(t, ja1.Validate())
	})
}

func TestJoinAuthNewAndSetAuth(t *testing.T) {
	ja := NewJoinAuth("jam").SetAuth("admin", "s00per")
	assert.Equal(t, ja.Intent(), Present)
	assert.Equal(t, ja.AuthID, "jam")
	assert.Equal(t, ja.Auth.Username, "admin")
	assert.Equal(t, ja.Auth.Password, "s00per")
	assert.Equal(t, ja.LinkedToCluster, "")
}

func TestJoinAuthNewToRemove(t *testing.T) {
	rc := NewJoinAuthToRemove("nope")
	assert.Equal(t, rc.Intent(), Removed)
	assert.Equal(t, rc.AuthID, "nope")
}

func TestJoinAuthNewLinked(t *testing.T) {
	c := NewActiveDirectoryCluster("c1", "test.example.net")
	ja := NewLinkedJoinAuth(c).SetAuth("admin", "3sc4l4t3")
	assert.NoError(t, ja.Validate())
	assert.Equal(t, ja.LinkedToCluster, "c1")
	assert.Len(t, c.DomainSettings.JoinSources, 1)
	assert.Equal(t, c.DomainSettings.JoinSources[0].Ref, ja.AuthID)
}

func TestJoinAuthMarshalUnmarshal(t *testing.T) {
	ja := NewJoinAuth("jjjj").SetAuth("admin", "3sc4l4t3")
	j, err := json.Marshal(ja)
	assert.NoError(t, err)
	ja2 := &JoinAuth{}
	err = json.Unmarshal(j, ja2)
	assert.NoError(t, err)
	assert.Equal(t, ja.AuthID, ja2.AuthID)
	assert.Equal(t, ja.Auth.Username, ja2.Auth.Username)
	assert.Equal(t, ja.Auth.Password, ja2.Auth.Password)
}

func TestJoinAuthSetAuth(t *testing.T) {
	ja := &JoinAuth{IntentValue: Present, AuthID: "abc"}
	ja.SetAuth("admin", "c00lt1m3")
	assert.NotNil(t, ja.Auth)
	assert.EqualValues(t, ja.Auth.Username, "admin")
	assert.EqualValues(t, ja.Auth.Password, "c00lt1m3")
}
