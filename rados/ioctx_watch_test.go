package rados_test

import (
	"testing"
	"sync"

	"github.com/ceph/go-ceph/rados"
	"github.com/stretchr/testify/assert"
)

var (
	mu           sync.Mutex
	shared_cache map[int][]byte
)

type TestWatcher struct {
	oid      string
	index    int
	ioctx    *rados.IOContext
}

func NewTestWatcher(oid string, index int, ioctx *rados.IOContext) *TestWatcher {
	return &TestWatcher{
		oid:   oid,
		index: index,
		ioctx: ioctx,
	}
}

func (w *TestWatcher) OnNotify(notify_id uint64, cookie uint64, notifier_id uint64, data []byte) {
	mu.Lock()
	defer mu.Unlock()

	shared_cache[w.index] = data

	w.ioctx.NotifyAck(w.oid, notify_id, cookie)
}

func (w *TestWatcher) OnError(cookie uint64, err int) {
}

func TestWaitNotify(t *testing.T) {
	shared_cache = map[int][]byte{}

	conn, _ := rados.NewConn()
	conn.ReadDefaultConfigFile()
	conn.Connect()

	// make pool
	pool_name := GetUUID()
	err := conn.MakePool(pool_name)
	assert.NoError(t, err)

	pool, err := conn.OpenIOContext(pool_name)
	assert.NoError(t, err)

	key := "obj"
	bytes_in := []byte("input data")
	err = pool.Write(key, bytes_in, 0)
	assert.NoError(t, err)

	watchers := 100
	var cookies []uint64
	for i := 0; i < watchers; i++ {
		watcher := NewTestWatcher(key, i, pool)
		cookie, err := pool.Watch(key, watcher)
		assert.NoError(t, err)
		cookies = append(cookies, cookie)
	}

	notify_in := []byte("notify data")
	err = pool.Notify(key, notify_in, 1000*5)
	assert.NoError(t, err)

	assert.Equal(t, len(shared_cache), watchers)

	shared_cache = map[int][]byte{}

	for _, cookie := range cookies {
		ret, err := pool.WatchCheck(cookie)
		assert.NoError(t, err)
		assert.Equal(t, ret > 0, true)

		err = pool.UnWatch(cookie)
		assert.NoError(t, err)
	}

	err = pool.Notify(key, notify_in, 1000*5)
	assert.NoError(t, err)

	assert.Equal(t, len(shared_cache), 0)
}
