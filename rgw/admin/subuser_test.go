package admin

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateSubuserAccess(t *testing.T) {
	assert.True(t, SubuserSpec{Access: SubuserAccessNone}.validateSubuserAccess())
	assert.True(t, SubuserSpec{Access: SubuserAccessRead}.validateSubuserAccess())
	assert.True(t, SubuserSpec{Access: SubuserAccessWrite}.validateSubuserAccess())
	assert.True(t, SubuserSpec{Access: SubuserAccessReadWrite}.validateSubuserAccess())
	assert.True(t, SubuserSpec{Access: SubuserAccessFull}.validateSubuserAccess())
	assert.False(t, SubuserSpec{Access: SubuserAccessReplyFull}.validateSubuserAccess())
	assert.False(t, SubuserSpec{Access: SubuserAccess("bar")}.validateSubuserAccess())
}

func TestCreateSubuserMockAPI(t *testing.T) {
	t.Run("test create subuser validation: neither is set", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClientCreateSubuser())
		assert.NoError(t, err)
		err = api.CreateSubuser(context.TODO(), User{}, SubuserSpec{})
		assert.Equal(t, err, errMissingUserID)
	})
	t.Run("test create subuser validation: no user ID", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClientCreateSubuser())
		assert.NoError(t, err)
		err = api.CreateSubuser(context.TODO(), User{}, SubuserSpec{Name: "foo"})
		assert.Equal(t, err, errMissingUserID)
	})
	t.Run("test create subuser validation: no subuser ID", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClientCreateSubuser())
		assert.NoError(t, err)
		err = api.CreateSubuser(context.TODO(), User{ID: "dashboard-admin"}, SubuserSpec{})
		assert.Equal(t, err, errMissingSubuserID)
	})
	t.Run("test create subuser validation: valid", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClientCreateSubuser())
		assert.NoError(t, err)
		err = api.CreateSubuser(context.TODO(), User{ID: "dashboard-admin"}, SubuserSpec{Name: "foo"})
		assert.NoError(t, err)
	})
	t.Run("test create subuser validation: invalid access", func(t *testing.T) {
		api, err := New("127.0.0.1", "accessKey", "secretKey", returnMockClientCreateSubuser())
		assert.NoError(t, err)
		err = api.CreateSubuser(context.TODO(), User{ID: "dashboard-admin"}, SubuserSpec{Name: "foo", Access: SubuserAccess("foo")})
		assert.Error(t, err)
		assert.EqualError(t, err, `invalid subuser access level "foo"`)
	})
}

// mockClient is defined in user_test.go
func returnMockClientCreateSubuser() *mockClient {
	r := ioutil.NopCloser(bytes.NewReader([]byte("")))
	return &mockClient{
		mockDo: func(req *http.Request) (*http.Response, error) {
			if req.Method == http.MethodPut && req.URL.Path == "127.0.0.1/admin/user" {
				return &http.Response{
					StatusCode: 201,
					Body:       r,
				}, nil
			}
			return nil, fmt.Errorf("unexpected request: %q. method %q. path %q", req.URL.RawQuery, req.Method, req.URL.Path)
		},
	}
}
