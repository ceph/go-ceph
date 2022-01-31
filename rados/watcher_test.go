//go:build ceph_preview
// +build ceph_preview

package rados

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func (suite *RadosTestSuite) TestWatcher() {
	suite.SetupConnection()
	oid := suite.GenObjectName()
	err := suite.ioctx.Create(oid, CreateExclusive)
	assert.NoError(suite.T(), err)
	defer func() { _ = suite.ioctx.Delete(oid) }()

	suite.T().Run("DoubleDelete", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		assert.NoError(t, err)
		assert.NotPanics(t, func() {
			err := watcher.Delete()
			assert.NoError(t, err)
			err = watcher.Delete()
			assert.NoError(t, err)
		})
	})

	suite.T().Run("DeleteClosesChannels", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		assert.NoError(t, err)
		evDone := make(chan struct{})
		errDone := make(chan struct{})
		go func() { // event handler
			for ne := range watcher.Events() {
				t.Errorf("received event: %v", ne)
			}
			close(evDone)
		}()
		go func() { // error handler
			for err := range watcher.Errors() {
				t.Errorf("received error: %v", err)
			}
			close(errDone)
		}()
		err = watcher.Delete()
		assert.NoError(t, err)
		select {
		case <-evDone:
			// Delete closed the events channel
		case <-time.After(time.Second):
			t.Error("timeout closing event channel")
		}
		select {
		case <-errDone:
			// Delete closed the error channel
		case <-time.After(time.Second):
			t.Error("timeout closing error channel")
		}
	})

	suite.T().Run("NotifyNoAck", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		defer func() { _ = watcher.Delete() }()
		assert.NoError(t, err)
		data := []byte("notification")
		acks, timeouts, err := suite.ioctx.NotifyWithTimeout(oid, data, time.Second)
		assert.Error(t, err) // without ack it must timeout
		assert.Len(t, timeouts, 1)
		assert.Equal(t, watcher.ID(), timeouts[0].WatcherID)
		assert.Empty(t, acks)
		select {
		case ev := <-watcher.Events():
			assert.Equal(t, data, ev.Data)
			assert.Equal(t, timeouts[0].NotifierID, ev.NotifierID)
		case err = <-watcher.Errors():
			t.Error(err)
		case <-time.After(time.Second):
			t.Error("timeout")
		}
	})

	suite.T().Run("NotifyDeletedWatcher", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		assert.NoError(t, err)
		err = watcher.Delete()
		assert.NoError(t, err)
		data := []byte("notification")
		acks, timeouts, err := suite.ioctx.NotifyWithTimeout(oid, data, time.Second)
		assert.NoError(t, err)
		assert.Empty(t, timeouts)
		assert.Empty(t, acks)
		select {
		case _, ok := <-watcher.Events():
			assert.False(t, ok)
		case <-time.After(time.Second):
			t.Error("timeout")
		}
	})

	suite.T().Run("NotifyNoAckDeleteUnblocksChannels", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		assert.NoError(t, err)
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			i := i
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _, err := suite.ioctx.Notify(oid, []byte(fmt.Sprintf("notify%d", i)))
				assert.NoError(t, err)
			}()
		}
		time.Sleep(time.Millisecond * 100) // wait so that callbacks are blocking
		err = watcher.Delete()
		assert.NoError(t, err)
		wg.Wait()
	})

	suite.T().Run("NotifyWithAck", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		defer func() { _ = watcher.Delete() }()
		assert.NoError(t, err)
		data := []byte("notification")
		response := []byte("response")
		var receivedEv NotifyEvent
		go func() { // event handler
			for ne := range watcher.Events() {
				receivedEv = ne
				assert.Equal(t, data, ne.Data)
				assert.Equal(t, watcher.ID(), ne.WatcherID)
				err := ne.Ack(response)
				assert.NoError(t, err)
			}
		}()
		go func() { // error handler
			for err := range watcher.Errors() {
				t.Errorf("received error: %v", err)
			}
		}()
		acks, timeouts, err := suite.ioctx.Notify(oid, data)
		assert.NoError(t, err) // without ack it must timeout
		assert.Empty(t, timeouts)
		assert.Len(t, acks, 1)
		assert.Equal(t, watcher.ID(), acks[0].WatcherID)
		assert.Equal(t, receivedEv.NotifierID, acks[0].NotifierID)
		assert.Equal(t, response, acks[0].Response)
	})

	suite.T().Run("NotifyAckNilData", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		defer func() { _ = watcher.Delete() }()
		assert.NoError(t, err)
		var receivedEv NotifyEvent
		go func() { // event handler
			for ne := range watcher.Events() {
				receivedEv = ne
				assert.Equal(t, []byte(nil), ne.Data)
				assert.Equal(t, watcher.ID(), ne.WatcherID)
				err := ne.Ack(nil)
				assert.NoError(t, err)
			}
		}()
		go func() { // error handler
			for err := range watcher.Errors() {
				t.Errorf("received error: %v", err)
			}
		}()
		acks, timeouts, err := suite.ioctx.Notify(oid, nil)
		assert.NoError(t, err) // without ack it must timeout
		assert.Empty(t, timeouts)
		assert.Len(t, acks, 1)
		assert.Equal(t, watcher.ID(), acks[0].WatcherID)
		assert.Equal(t, receivedEv.NotifierID, acks[0].NotifierID)
		assert.Equal(t, []byte(nil), acks[0].Response)
	})

	suite.T().Run("Check", func(t *testing.T) {
		watcher, err := suite.ioctx.WatchWithTimeout(oid, time.Second)
		defer func() { _ = watcher.Delete() }()
		assert.NoError(t, err)
		last, err := watcher.Check()
		assert.NoError(t, err)
		assert.Greater(t, int(last), 0)
		select {
		case err = <-watcher.Errors(): // watcher times out
		case <-time.After(time.Second * 2):
			t.Error("timeout")
		}
		assert.Error(t, err)
		last, err2 := watcher.Check()
		assert.Error(t, err2)
		assert.Equal(t, err, err2)
		assert.Zero(t, last)
	})

	suite.T().Run("Flush", func(t *testing.T) {
		watcher, err := suite.ioctx.Watch(oid)
		assert.NoError(t, err)
		done := make(chan struct{})
		go func() {
			_, _, _ = suite.ioctx.Notify(oid, nil)
			close(done)
		}()
		time.Sleep(time.Millisecond * 100)
		flushed := make(chan struct{})
		go func() {
			err := suite.conn.WatcherFlush()
			assert.NoError(t, err)
			close(flushed)
		}()
		select {
		case <-flushed:
			t.Error("flush returned before event got received")
		case <-time.After(time.Millisecond * 100):
		}
		<-watcher.Events()
		select {
		case <-flushed:
		case <-time.After(time.Second):
			t.Error("flush didn't return after receiving event")
		}
		err = watcher.Delete()
		assert.NoError(t, err)
		<-done
	})
}

func TestDecodeNotifyResponse(t *testing.T) {
	testEmptyResponse := [...]byte{
		//    le32 num_acks
		0, 0, 0, 0,
		//    le32 num_timeouts
		0, 0, 0, 0,
	}
	testResponse := [...]byte{
		//    le32 num_acks
		1, 0, 0, 0,
		//      le64 gid     global id for the client (for client.1234 that's 1234)
		0, 1, 0, 0, 0, 0, 0, 0,
		//      le64 cookie  cookie for the client
		1, 1, 0, 0, 0, 0, 0, 0,
		//      le32 buflen  length of reply message buffer
		4, 0, 0, 0,
		//      u8 buflen  payload
		1, 2, 3, 4,
		//    le32 num_timeouts
		2, 0, 0, 0,
		//      le64 gid     global id for the client
		2, 1, 0, 0, 0, 0, 0, 0,
		//      le64 cookie  cookie for the client
		3, 1, 0, 0, 0, 0, 0, 0,
		//      le64 gid     global id for the client
		4, 1, 0, 0, 0, 0, 0, 0,
		//      le64 cookie  cookie for the client
		255, 255, 255, 255, 255, 255, 255, 255,
	}
	t.Run("Empty", func(t *testing.T) {
		l := _Ctype_size_t(len(testEmptyResponse))
		b := (*_Ctype_char)(unsafe.Pointer(&testEmptyResponse[0]))
		acks, tOuts := decodeNotifyResponse(b, l)
		assert.Len(t, acks, 0)
		assert.Len(t, tOuts, 0)
	})
	t.Run("Example", func(t *testing.T) {
		l := _Ctype_size_t(len(testResponse))
		b := (*_Ctype_char)(unsafe.Pointer(&testResponse[0]))
		acks, tOuts := decodeNotifyResponse(b, l)
		assert.Len(t, acks, 1)
		assert.Equal(t, acks[0].NotifierID, NotifierID(256))
		assert.Equal(t, acks[0].WatcherID, WatcherID(257))
		assert.Len(t, acks[0].Response, 4)
		assert.Equal(t, acks[0].Response, []byte{1, 2, 3, 4})
		assert.Len(t, tOuts, 2)
		assert.Equal(t, tOuts[0].NotifierID, NotifierID(258))
		assert.Equal(t, tOuts[0].WatcherID, WatcherID(259))
		assert.Equal(t, tOuts[1].NotifierID, NotifierID(260))
		assert.Equal(t, tOuts[1].WatcherID, WatcherID(math.MaxUint64))
	})
}
