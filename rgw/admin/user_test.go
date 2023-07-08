package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
  "subusers": [
     {
        "id": "dashboard-admin:swift",
        "permissions": "read"
     }
  ],
  "keys": [
    {
      "user": "dashboard-admin",
      "access_key": "4WD1FGM5PXKLC97YC0SZ",
      "secret_key": "YSaT5bEcJTjBJCDG5yvr2NhGQ9xzoTIg8B1gQHa3"
    }
  ],
  "swift_keys": [
    {
      "user": "dashboard-admin:swift",
      "secret_key": "VERY_SECRET"
    }
  ],
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

	suite.T().Run("get user leseb by uid", func(t *testing.T) {
		user, err := co.GetUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@example.com", user.Email)
		assert.Equal(suite.T(), "users", user.Caps[0].Type)
		assert.Equal(suite.T(), "read", user.Caps[0].Perm)
		os.Setenv("LESEB_ACCESS_KEY", user.Keys[0].AccessKey)
	})

	suite.T().Run("get user leseb by key", func(t *testing.T) {
		user, err := co.GetUser(context.Background(), User{Keys: []UserKeySpec{{AccessKey: os.Getenv("LESEB_ACCESS_KEY")}}})
		assert.NoError(suite.T(), err)
		assert.Equal(suite.T(), "leseb@example.com", user.Email)
		os.Unsetenv("LESEB_ACCESS_KEY")
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

	suite.T().Run("create a subuser", func(t *testing.T) {
		err := co.CreateSubuser(context.Background(), User{ID: "leseb"}, SubuserSpec{Name: "foo", Access: SubuserAccessReadWrite})
		assert.NoError(suite.T(), err)

		user, err := co.GetUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
		if err == nil {
			assert.Equal(suite.T(), user.Subusers[0].Name, "leseb:foo")
			// Note: the returned values are not equal to the input values ...
			assert.Equal(suite.T(), user.Subusers[0].Access, SubuserAccess("read-write"))
		}
	})

	suite.T().Run("modify a subuser", func(t *testing.T) {
		err := co.ModifySubuser(context.Background(), User{ID: "leseb"}, SubuserSpec{Name: "foo", Access: SubuserAccessRead})
		assert.NoError(suite.T(), err)

		user, err := co.GetUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
		if err == nil {
			assert.Equal(suite.T(), user.Subusers[0].Name, "leseb:foo")
			assert.Equal(suite.T(), user.Subusers[0].Access, SubuserAccess("read"))
		}
	})

	suite.T().Run("remove a subuser", func(t *testing.T) {
		err := co.RemoveSubuser(context.Background(), User{ID: "leseb"}, SubuserSpec{Name: "foo"})
		assert.NoError(suite.T(), err)

		user, err := co.GetUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
		if err == nil {
			assert.Equal(suite.T(), len(user.Subusers), 0)
		}
	})

	suite.T().Run("remove user", func(t *testing.T) {
		err = co.RemoveUser(context.Background(), User{ID: "leseb"})
		assert.NoError(suite.T(), err)
	})
}

func TestGetUserMockAPI(t *testing.T) {
	t.Run("test simple api mock", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClient())
		assert.NoError(t, err)
		u, err := api.GetUser(context.TODO(), User{ID: "dashboard-admin"})
		assert.NoError(t, err)
		assert.Equal(t, "dashboard-admin", u.DisplayName, u)
	})
	t.Run("test get user with access key", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClient())
		assert.NoError(t, err)
		u, err := api.GetUser(context.TODO(), User{Keys: []UserKeySpec{{AccessKey: "4WD1FGM5PXKLC97YC0SZ"}}})
		assert.NoError(t, err)
		assert.Equal(t, "dashboard-admin", u.DisplayName, u)
	})
	t.Run("test get user with nothing", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClient())
		assert.NoError(t, err)
		_, err = api.GetUser(context.TODO(), User{})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing user ID")
	})
	t.Run("test get user with missing correct key", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClient())
		assert.NoError(t, err)
		_, err = api.GetUser(context.TODO(), User{Keys: []UserKeySpec{{SecretKey: "4WD1FGM5PXKLC97YC0SZ"}}})
		assert.Error(t, err)
		assert.EqualError(t, err, "missing user access key")
	})
}

func returnMockClient() *mockClient {
	r := io.NopCloser(bytes.NewReader(fakeUserResponse))
	return &mockClient{
		mockDo: func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodGet && req.URL.Path == "127.0.0.1/admin/user" {
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			}
			return nil, fmt.Errorf("unexpected request: %q. method %q. path %q", req.URL.RawQuery, req.Method, req.URL.Path)
		},
	}
}
