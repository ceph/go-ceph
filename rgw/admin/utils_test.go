package admin

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getDefaultValue() url.Values {
	values := url.Values{}
	values.Add("format", "json")

	return values
}

func TestBuildQueryPath(t *testing.T) {
	queryPath := buildQueryPath("http://192.168.0.1", "/user", getDefaultValue().Encode())
	assert.Equal(t, "http://192.168.0.1/admin/user?format=json", queryPath)

	queryPath = buildQueryPath("http://192.168.0.1", "/user?foo", getDefaultValue().Encode())
	assert.Equal(t, "http://192.168.0.1/admin/user?foo&format=json", queryPath)
}

func TestValueToURLParams(t *testing.T) {
	type args struct {
		i                interface{}
		acceptableFields []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{User{ID: "leseb", Email: "leseb@example.com"}, []string{"uid"}}, "format=json&uid=leseb"},
		// RGW expects placement-tags as a single comma-separated value, not repeated params.
		{"placement tags", args{User{ID: "leseb", PlacementTags: []string{"fast", "ssd"}}, []string{"uid", "placement-tags"}}, "format=json&placement-tags=fast%2Cssd&uid=leseb"},
		{"empty placement tags", args{User{ID: "leseb"}, []string{"uid", "placement-tags"}}, "format=json&uid=leseb"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := valueToURLParams(tt.args.i, tt.args.acceptableFields)
			if !reflect.DeepEqual(got.Encode(), tt.want) {
				t.Errorf("valueToURLParams() = %v, want %v", got.Encode(), tt.want)
			}
		})
	}
}
