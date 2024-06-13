package admin

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	fakeGetUserError    = []byte(`{"Code":"NoSuchUser","RequestId":"tx0000000000000000005a9-00608957a2-10496-my-store","HostId":"10496-my-store-my-store"}`)
	fakeGetSubUserError = []byte(`{"Code":"NoSuchSubUser","RequestId":"tx0000000000000000005a9-00608957a2-10496-my-store","HostId":"10496-my-store-my-store"}`)
)

func TestHandleStatusError(t *testing.T) {
	err := handleStatusError(fakeGetUserError)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSuchUser), err)

	err = handleStatusError(fakeGetSubUserError)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrNoSuchSubUser), err)
}
