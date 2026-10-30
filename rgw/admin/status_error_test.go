//go:build ceph_preview

package admin

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetUserStatusError verifies that the HTTP status code of a failed admin
// request propagates through call() and can be retrieved via errors.As, even
// when the gateway reports a generic "UnknownError" (HTTP 500).
func TestGetUserStatusError(t *testing.T) {
	mc := &mockClient{
		mockDo: func(_ *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(bytes.NewReader([]byte(`{"Code":"UnknownError"}`))),
			}, nil
		},
	}
	api, err := New("127.0.0.1", "accessKey", "secretKey", mc)
	assert.NoError(t, err)

	_, err = api.GetUser(context.TODO(), User{ID: "some-user"})
	assert.Error(t, err)

	var se StatusError
	assert.True(t, errors.As(err, &se), err)
	assert.Equal(t, http.StatusInternalServerError, se.StatusCode)
	// The error code is not a public field, but it still surfaces in the message.
	assert.Contains(t, se.Error(), "UnknownError")
}
