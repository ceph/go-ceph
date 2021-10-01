//go:build !nautilus && ceph_preview
// +build !nautilus,ceph_preview

package admin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewImageSpec(t *testing.T) {
	type args struct {
		pool      string
		namespace string
		image     string
	}
	tests := []struct {
		name string
		args args
		want ImageSpec
	}{
		{
			name: "onlyImageName",
			args: args{
				pool:      "",
				namespace: "",
				image:     "img",
			},
			want: ImageSpec{
				spec: "img",
			},
		},
		{
			name: "Image&PoolName",
			args: args{
				pool:      "pool",
				namespace: "",
				image:     "img",
			},
			want: ImageSpec{
				spec: "pool/img",
			},
		},
		{
			name: "all args",
			args: args{
				pool:      "pool",
				namespace: "ns",
				image:     "img",
			},
			want: ImageSpec{
				spec: "pool/ns/img",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewImageSpec(tt.args.pool, tt.args.namespace, tt.args.image))
		})
	}
}

func TestNewRawImageSpec(t *testing.T) {
	type args struct {
		spec string
	}
	tests := []struct {
		name string
		args args
		want ImageSpec
	}{
		{
			name: "valid",
			args: args{
				spec: "pool/img",
			},
			want: ImageSpec{
				spec: "pool/img",
			},
		},
		{
			name: "invalid but still accepts",
			args: args{
				spec: "invalid format ...",
			},
			want: ImageSpec{
				spec: "invalid format ...",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NewRawImageSpec(tt.args.spec))
		})
	}
}
