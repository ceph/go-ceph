package admin

import (
	"errors"
	"fmt"
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
func TestParseError(t *testing.T) {
	var (
		nilError         error = nil
		nilErrorExpected       = statusError{}

		nonValidStatusError         = errors.New("missing user ID")
		nonValidStatusErrorExpected = statusError{}

		getUserError error = statusError{
			Code:      "NoSuchUser",
			RequestID: "tx0000000000000000005a9-00608957a2-10496-my-store",
			HostID:    "10496-my-store-my-store",
		}
		getUserErrorExpected = statusError{
			Code:      "NoSuchUser",
			RequestID: "tx0000000000000000005a9-00608957a2-10496-my-store",
			HostID:    "10496-my-store-my-store",
		}

		getSubUserError error = statusError{
			Code:      "NoSuchSubUser",
			RequestID: "tx0000000000000000005a9-00608957a2-10496-my-store",
			HostID:    "10496-my-store-my-store",
		}
		getSubUserErrorExpected = statusError{
			Code:      "NoSuchSubUser",
			RequestID: "tx0000000000000000005a9-00608957a2-10496-my-store",
			HostID:    "10496-my-store-my-store",
		}
	)

	// Nil error
	statusError, ok := ParseError(nilError)
	assert.True(t, ok)
	assert.Equal(t, nilErrorExpected, statusError)

	// Non-valid error
	statusError, ok = ParseError(nonValidStatusError)
	assert.False(t, ok)
	assert.Equal(t, nonValidStatusErrorExpected, statusError)

	// Get user error
	statusError, ok = ParseError(getUserError)
	assert.True(t, ok)
	assert.EqualExportedValues(t, getUserErrorExpected, statusError)
	assert.True(t, statusError.Is(ErrNoSuchUser))

	// Get sub-user error
	statusError, ok = ParseError(getSubUserError)
	assert.True(t, ok)
	assert.EqualExportedValues(t, getSubUserErrorExpected, statusError)
	assert.True(t, statusError.Is(ErrNoSuchSubUser))

	// Wrapped error
	wrappedError := fmt.Errorf("additional note: %w", getUserError)
	statusError, ok = ParseError(wrappedError)
	assert.True(t, ok)
	assert.EqualExportedValues(t, getUserErrorExpected, statusError)
	assert.True(t, statusError.Is(ErrNoSuchUser))
}
