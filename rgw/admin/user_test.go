package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockClient is the mock of the HTTP Client
// It can be used to mock HTTP request/response from the rgw admin ops API
type mockClient struct {
	// mockDo is a type that mock the Do method from the HTTP package
	mockDo mockDoType
}

// mockDoType is a custom type that allows setting the function that our Mock Do func will run instead
type mockDoType func(req *http.Request) (*http.Response, error)

// Do is the mock client's `Do` func
func (m *mockClient) Do(req *http.Request) (*http.Response, error) { return m.mockDo(req) }

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
	co, err := New(suite.endpoint, suite.accessKey, suite.secretKey, newDebugHTTPClient(http.DefaultClient))
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
		usercaps := "users=read"
		user, err := co.CreateUser(context.Background(), User{ID: "leseb", DisplayName: "This is leseb", Email: "leseb@example.com", UserCaps: usercaps})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@example.com", user.Email)
	})

	suite.T().Run("get user leseb", func(t *testing.T) {
		user, err := co.GetUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@example.com", user.Email)
		assert.Equal(suite.T(), "users", user.Caps[0].Type)
		assert.Equal(suite.T(), "read", user.Caps[0].Perm)
	})

	suite.T().Run("modify user email", func(t *testing.T) {
		user, err := co.ModifyUser(context.Background(), User{ID: "leseb", Email: "leseb@leseb.com"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@leseb.com", user.Email)
	})

	suite.T().Run("modify user max bucket", func(t *testing.T) {
		maxBuckets := -1
		user, err := co.ModifyUser(context.Background(), User{ID: "leseb", MaxBuckets: &maxBuckets})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@leseb.com", user.Email)
		assert.Equal(suite.T(), -1, *user.MaxBuckets)
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

	suite.T().Run("set user quota", func(t *testing.T) {
		quotaEnable := true
		maxObjects := int64(100)
		err := co.SetUserQuota(context.Background(), QuotaSpec{QuotaType: "user", UID: "leseb", MaxObjects: &maxObjects, Enabled: &quotaEnable})
		assert.NoError(suite.T(), err)
	})

	suite.T().Run("get user quota", func(t *testing.T) {
		q, err := co.GetUserQuota(context.Background(), QuotaSpec{QuotaType: "user", UID: "leseb"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), int64(100), *q.MaxObjects)
	})

	suite.T().Run("get user stat", func(t *testing.T) {
		statEnable := true
		user, err := co.GetUser(context.Background(), User{ID: "leseb", GenerateStat: &statEnable})
		assert.NoError(suite.T(), err)
		assert.NotNil(suite.T(), user.Stat.Size)
	})

	suite.T().Run("remove user", func(t *testing.T) {
		err = co.RemoveUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
	})
}

func TestGetUserMockAPI(t *testing.T) {
	r := ioutil.NopCloser(bytes.NewReader(fakeUserResponse))
	mockClient := &mockClient{
		mockDo: func(req *http.Request) (*http.Response, error) {
			if req.URL.RawQuery == "format=json&uid=dashboard-admin" && req.Method == http.MethodGet && req.URL.Path == "127.0.0.1/admin/user" {
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			}
			return nil, fmt.Errorf("unexpected request: %q. method %q. path %q", req.URL.RawQuery, req.Method, req.URL.Path)
		},
	}

	api, err := New("127.0.0.1", "accessKey", "secretKey", mockClient)
	assert.NoError(t, err)
	u, err := api.GetUser(context.TODO(), User{ID: "dashboard-admin"})
	assert.NoError(t, err)
	assert.Equal(t, "dashboard-admin", u.DisplayName, u)
}
