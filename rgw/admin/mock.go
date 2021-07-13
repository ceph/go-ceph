package admin

import (
	"net/http"
)

// MockClient is the mock of the HTTP Client
// It can be used to mock HTTP request/response from the rgw admin ops API
type MockClient struct {
	// MockDo is a type that mock the Do method from the HTTP package
	MockDo MockDoType
}

// MockDoType is a custom type that allows setting the function that our Mock Do func will run instead
type MockDoType func(req *http.Request) (*http.Response, error)

// Do is the mock client's `Do` func
func (m *MockClient) Do(req *http.Request) (*http.Response, error) { return m.MockDo(req) }
