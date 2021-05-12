package admin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	fakeUserResponse = []byte(`
{
  "tenant": "",
  "user_id": "dashboard-admin",
  "display_name": "dashboard-admin",
  "email": "",
  "suspended": 0,
  "max_buckets": 1000,
  "subusers": [],
  "keys": [
    {
      "user": "dashboard-admin",
      "access_key": "4WD1FGM5PXKLC97YC0SZ",
      "secret_key": "YSaT5bEcJTjBJCDG5yvr2NhGQ9xzoTIg8B1gQHa3"
    }
  ],
  "swift_keys": [],
  "caps": [],
  "op_mask": "read, write, delete",
  "system": "true",
  "admin": "false",
  "default_placement": "",
  "default_storage_class": "",
  "placement_tags": [],
  "bucket_quota": {
    "enabled": false,
    "check_on_raw": false,
    "max_size": -1,
    "max_size_kb": 0,
    "max_objects": -1
  },
  "user_quota": {
    "enabled": false,
    "check_on_raw": false,
    "max_size": -1,
    "max_size_kb": 0,
    "max_objects": -1
  },
  "temp_url_keys": [],
  "type": "rgw",
  "mfa_ids": []
}`)
)

func TestUnmarshal(t *testing.T) {
	u := &User{}
	err := json.Unmarshal(fakeUserResponse, &u)
	assert.NoError(t, err)
}

func (suite *RadosGWTestSuite) TestUser() {
	suite.SetupConnection()
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, nil)
	co.Debug = true
	assert.NoError(suite.T(), err)

	suite.T().Run("fail to create user since no UID provided", func(t *testing.T) {
		_, err = co.CreateUser(context.Background(), User{Email: "leseb@example.com"})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserID.Error())
	})

	suite.T().Run("fail to create user since no no display name provided", func(t *testing.T) {
		_, err = co.CreateUser(context.Background(), User{ID: "leseb", Email: "leseb@example.com"})
		assert.Error(suite.T(), err)
		assert.EqualError(suite.T(), err, errMissingUserDisplayName.Error())
	})

	suite.T().Run("user creation success", func(t *testing.T) {
		user, err := co.CreateUser(context.Background(), User{ID: "leseb", DisplayName: "This is leseb", Email: "leseb@example.com"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@example.com", user.Email)
	})

	suite.T().Run("get user leseb", func(t *testing.T) {
		user, err := co.GetUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@example.com", user.Email)
	})

	suite.T().Run("modify user email", func(t *testing.T) {
		user, err := co.ModifyUser(context.Background(), User{ID: "leseb", Email: "leseb@leseb.com"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@leseb.com", user.Email)
	})

	suite.T().Run("user already exists", func(t *testing.T) {
		_, err := co.CreateUser(context.Background(), User{ID: "admin", DisplayName: "Admin user"})
		assert.Error(suite.T(), err)
		assert.True(suite.T(), errors.Is(err, ErrUserExists), fmt.Sprintf("%+v", err))
	})

	suite.T().Run("get users", func(t *testing.T) {
		users, err := co.GetUsers(context.Background())
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), 2, len(*users))
	})

	suite.T().Run("remove user", func(t *testing.T) {
		err = co.RemoveUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
	})
}
