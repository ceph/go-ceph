//go:build !nautilus && ceph_preview
// +build !nautilus,ceph_preview

package admin

import (
	"errors"
	"testing"
	"time"

	"github.com/ceph/go-ceph/internal/commands"
	"github.com/ceph/go-ceph/rbd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var tr1 = `
{
   "sequence":1,
   "id":"id-1",
   "message":"Removing image pool/test from trash",
   "refs":{
      "action":"trash remove",
      "pool_name":"pool",
      "pool_namespace":"",
      "image_id":"12345"
   },
   "retry_attempts":1,
   "retry_time":"2021-09-13T05:58:24.826408"
}
`

var tr2 = `
{
	"sequence":2,
	"id":"id-2",
	"message":"Removing image pool/test",
	"refs":{
	"action":"remove",
	"pool_name":"pool",
	"pool_namespace":"",
	"image_name":"test",
	"image_id":"123456"
	},
	"in_progress":true,
	"progress":0.70
}
`

var trList = `[
	{
		"sequence":1,
		"id":"id-1",
		"message":"Removing image pool/test from trash",
		"refs":{
		   "action":"trash remove",
		   "pool_name":"pool",
		   "pool_namespace":"",
		   "image_id":"12345"
		},
		"retry_attempts":1,
		"retry_time":"2021-09-13T05:58:24.826408"
	 },
	 {
		"sequence":2,
		"id":"id-2",
		"message":"Removing image pool/test",
		"refs":{
		"action":"remove",
		"pool_name":"pool",
		"pool_namespace":"",
		"image_name":"test",
		"image_id":"123456"
		},
		"in_progress":true,
		"progress":0.70
	}
]`

func TestParseTaskResponse(t *testing.T) {
	type args struct {
		res commands.Response
	}
	tests := []struct {
		name    string
		args    args
		want    TaskResponse
		wantErr bool
	}{
		{
			name: "",
			args: args{
				res: commands.NewResponse([]byte(tr1), "", nil),
			},
			want: TaskResponse{
				Sequence: 1,
				ID:       "id-1",
				Message:  "Removing image pool/test from trash",
				Refs: TaskRefs{
					Action:        "trash remove",
					PoolName:      "pool",
					PoolNamespace: "",
					ImageName:     "",
					ImageID:       "12345",
				},
				InProgress:    false,
				Progress:      0,
				RetryAttempts: 1,
				RetryTime:     "2021-09-13T05:58:24.826408",
				RetryMessage:  "",
			},
			wantErr: false,
		},
		{
			name: "",
			args: args{
				res: commands.NewResponse([]byte(tr2), "", nil),
			},
			want: TaskResponse{
				Sequence: 2,
				ID:       "id-2",
				Message:  "Removing image pool/test",
				Refs: TaskRefs{
					Action:        "remove",
					PoolName:      "pool",
					PoolNamespace: "",
					ImageName:     "test",
					ImageID:       "123456",
				},
				InProgress:    true,
				Progress:      0.70,
				RetryAttempts: 0,
				RetryTime:     "",
				RetryMessage:  "",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTaskResponse(tt.args.res)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TesParseTaskResponseList(t *testing.T) {
	type args struct {
		res commands.Response
	}
	res1, err := parseTaskResponse(commands.NewResponse([]byte(tr1), "", nil))
	assert.NoError(t, err)
	res2, err := parseTaskResponse(commands.NewResponse([]byte(tr2), "", nil))
	assert.NoError(t, err)

	tests := []struct {
		name    string
		args    args
		want    []TaskResponse
		wantErr bool
	}{
		{
			name: "emptyList",
			args: args{
				res: commands.NewResponse([]byte(`[]`), "", nil),
			},
			want:    []TaskResponse{},
			wantErr: false,
		},
		{
			name: "twoItemList",
			args: args{
				res: commands.NewResponse([]byte(trList), "", nil),
			},
			want:    []TaskResponse{res1, res2},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTaskResponseList(tt.args.res)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

var (
	imageName = "img"
	testNS    = "test"
)

type args struct {
	pool      string
	namespace string
	imageName string
}

var tests = []struct {
	name string
	args args
}{
	{
		name: "onlyImageName",
		args: args{
			pool:      "",
			namespace: "",
			imageName: imageName,
		},
	},
	{
		name: "Image&PoolName",
		args: args{
			pool:      defaultPoolName,
			namespace: "",
			imageName: imageName,
		},
	},
	{
		name: "AllArgs",
		args: args{
			pool:      defaultPoolName,
			namespace: testNS,
			imageName: imageName,
		},
	},
}

func TestTaskAdminAddRemove(t *testing.T) {
	ensureDefaultPool(t)
	conn := getConn(t)

	ioctx, err := conn.OpenIOContext(defaultPoolName)
	require.NoError(t, err)
	defer ioctx.Destroy()

	assert.NoError(t, rbd.NamespaceCreate(ioctx, testNS))
	defer func() {
		assert.NoError(t, rbd.NamespaceRemove(ioctx, testNS))
	}()

	ta := getAdmin(t).Task()
	options := rbd.NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(rbd.ImageOptionOrder, uint64(testImageOrder)))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ioctx.SetNamespace(tt.args.namespace)
			err = rbd.CreateImage(ioctx, tt.args.imageName, testImageSize, options)
			assert.NoError(t, err)

			tr, err := ta.AddRemove(NewImageSpec(tt.args.pool, tt.args.namespace, tt.args.imageName))
			assert.NoError(t, err)

			assert.Equal(t, tt.args.imageName, tr.Refs.ImageName)
			assert.Equal(t, defaultPoolName, tr.Refs.PoolName)
			assert.Equal(t, tt.args.namespace, tr.Refs.PoolNamespace)
			assert.Equal(t, "remove", tr.Refs.Action)

			found := false
			// wait for the image to be deleted
			for i := 0; i < 35; i++ {
				imgList, err := rbd.GetImageNames(ioctx)
				assert.NoError(t, err)

				found = false
				for _, img := range imgList {
					if img == imageName {
						found = true
						break
					}
				}
				if !found {
					break
				}
				time.Sleep(time.Second)
			}
			assert.Equal(t, false, found)
		})
	}
}

func TestTaskAdminAddTrashRemove(t *testing.T) {
	ensureDefaultPool(t)
	conn := getConn(t)

	ioctx, err := conn.OpenIOContext(defaultPoolName)
	require.NoError(t, err)
	defer ioctx.Destroy()

	assert.NoError(t, rbd.NamespaceCreate(ioctx, testNS))
	defer func() {
		assert.NoError(t, rbd.NamespaceRemove(ioctx, testNS))
	}()

	ta := getAdmin(t).Task()
	options := rbd.NewRbdImageOptions()
	assert.NoError(t,
		options.SetUint64(rbd.ImageOptionOrder, uint64(testImageOrder)))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			ioctx.SetNamespace(tt.args.namespace)
			err = rbd.CreateImage(ioctx, tt.args.imageName, testImageSize, options)
			assert.NoError(t, err)

			img, err := rbd.OpenImage(ioctx, tt.args.imageName, rbd.NoSnapshot)
			assert.NoError(t, err)

			imageID, err := img.GetId()
			assert.NoError(t, err)

			assert.NoError(t, img.Trash(0))
			assert.NoError(t, img.Close())

			tr, err := ta.AddTrashRemove(NewImageSpec(tt.args.pool, tt.args.namespace, imageID))
			assert.NoError(t, err)

			assert.Equal(t, imageID, tr.Refs.ImageID)
			assert.Equal(t, defaultPoolName, tr.Refs.PoolName)
			assert.Equal(t, tt.args.namespace, tr.Refs.PoolNamespace)
			assert.Equal(t, "trash remove", tr.Refs.Action)

			trashList := []rbd.TrashInfo{}
			// wait for the image to be deleted
			for i := 0; i < 35; i++ {
				trashList, err = rbd.GetTrashList(ioctx)
				assert.NoError(t, err)

				if len(trashList) == 0 {
					break
				}
				time.Sleep(time.Second)
			}
			assert.Equal(t, 0, len(trashList))
		})
	}
}

func TestTaskAdminAddFlatten(t *testing.T) {
	parentImageName := "parent"

	ensureDefaultPool(t)
	conn := getConn(t)

	ioctx, err := conn.OpenIOContext(defaultPoolName)
	require.NoError(t, err)
	defer ioctx.Destroy()

	assert.NoError(t, rbd.NamespaceCreate(ioctx, testNS))
	defer func() {
		assert.NoError(t, rbd.NamespaceRemove(ioctx, testNS))
	}()

	ta := getAdmin(t).Task()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ioctx.SetNamespace(tt.args.namespace)

			options := rbd.NewRbdImageOptions()
			assert.NoError(t,
				options.SetUint64(rbd.ImageOptionOrder, uint64(testImageOrder)))
			err = rbd.CreateImage(ioctx, parentImageName, testImageSize, options)
			assert.NoError(t, err)

			parentImage, err := rbd.OpenImage(ioctx, parentImageName, rbd.NoSnapshot)
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, parentImage.Close())
				assert.NoError(t, parentImage.Remove())
			}()

			snap, err := parentImage.CreateSnapshot(tt.args.imageName)
			assert.NoError(t, err)

			err = snap.Protect()
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, snap.Unprotect())
				assert.NoError(t, snap.Remove())
			}()
			assert.NoError(t, options.SetUint64(rbd.ImageOptionFormat, uint64(2)))

			assert.NoError(t, rbd.CloneImage(ioctx, parentImageName, tt.args.imageName, ioctx, tt.args.imageName, options))

			childImage, err := rbd.OpenImage(ioctx, tt.args.imageName, rbd.NoSnapshot)
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, childImage.Close())
				assert.NoError(t, childImage.Remove())
			}()

			parentInfo, err := childImage.GetParent()
			assert.NoError(t, err)
			assert.Equal(t, parentImageName, parentInfo.Image.ImageName)
			assert.Equal(t, tt.args.imageName, parentInfo.Snap.SnapName)

			tr, err := ta.AddFlatten(NewImageSpec(tt.args.pool, tt.args.namespace, tt.args.imageName))
			assert.NoError(t, err)

			assert.Equal(t, tt.args.imageName, tr.Refs.ImageName)
			assert.Equal(t, defaultPoolName, tr.Refs.PoolName)
			assert.Equal(t, tt.args.namespace, tr.Refs.PoolNamespace)
			assert.Equal(t, "flatten", tr.Refs.Action)

			// wait for the image to be flattened
			for i := 0; i < 35; i++ {
				_, err = childImage.GetParent()
				if errors.Is(err, rbd.RbdErrorNotFound) {
					break
				}
				assert.NoError(t, err)

				time.Sleep(time.Second)
			}
			assert.Error(t, err)
		})
	}

}
