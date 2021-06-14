// +build !nautilus

package admin

import (
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ceph/go-ceph/internal/commands"
)

var ssList1 = `
{
    "4": {
        "name": "rbd/",
        "schedule": [
            {
                "interval": "70m",
                "start_time": null
            },
            {
                "interval": "30m",
                "start_time": null
            }
        ]
    },
    "4//106ff127efdc": {
        "name": "rbd/jumpy",
        "schedule": [
            {
                "interval": "99m",
                "start_time": null
            }
        ]
    }
}
`

var ssList2 = `
{
    "4": {
        "name": "rbd/",
        "schedule": [
            {
                "interval": "80m",
                "start_time": null
            }
        ]
    },
    "4//104f1d296736": {
        "name": "rbd/baz",
        "schedule": [
            {
                "interval": "1d",
                "start_time": "14:00:00-05:00"
            }
        ]
    }
}
`

var sStatus1 = `
{
    "scheduled_images": []
}
`
var sStatus2 = `
{
    "scheduled_images": [
        {
            "image": "rbd/foo",
            "schedule_time": "2021-03-02 16:30:00"
        }
    ]
}
`

var sStatus3 = `
{
    "scheduled_images": [
        {
            "image": "rbd/bar",
            "schedule_time": "2021-03-02 16:00:00"
        },
        {
            "image": "rbd/foo",
            "schedule_time": "2021-03-02 16:30:00"
        }
    ]
}
`

func TestParseMirrorSnapshotScheduleList(t *testing.T) {
	t.Run("list1", func(t *testing.T) {
		r := commands.NewResponse([]byte(ssList1), "", nil)
		l, err := parseMirrorSnapshotScheduleList(r)
		assert.NoError(t, err)
		if assert.Len(t, l, 2) {
			s1 := l[0]
			s2 := l[1]
			if s1.Name != "rbd/" {
				// just swap them.  it shouldn't matter to the test if the map has
				// changed the order.
				s1, s2 = s2, s1
			}
			assert.Equal(t, "rbd/", s1.Name)
			assert.Equal(t, "4", s1.LevelSpecID)
			if assert.Len(t, s1.Schedule, 2) {
				assert.EqualValues(t, "70m", s1.Schedule[0].Interval)
				assert.EqualValues(t, "", s1.Schedule[0].StartTime)
				assert.EqualValues(t, "30m", s1.Schedule[1].Interval)
				assert.EqualValues(t, "", s1.Schedule[1].StartTime)
			}

			assert.Equal(t, "rbd/jumpy", s2.Name)
			assert.Equal(t, "4//106ff127efdc", s2.LevelSpecID)
			if assert.Len(t, s2.Schedule, 1) {
				assert.EqualValues(t, "99m", s2.Schedule[0].Interval)
				assert.EqualValues(t, "", s2.Schedule[0].StartTime)
			}
		}
	})
	t.Run("list2", func(t *testing.T) {
		r := commands.NewResponse([]byte(ssList2), "", nil)
		l, err := parseMirrorSnapshotScheduleList(r)
		assert.NoError(t, err)
		if assert.Len(t, l, 2) {
			s1 := l[0]
			s2 := l[1]
			if s1.Name != "rbd/" {
				// just swap them.  it shouldn't matter to the test if the map has
				// changed the order.
				s1, s2 = s2, s1
			}
			assert.Equal(t, "rbd/", s1.Name)
			assert.Equal(t, "4", s1.LevelSpecID)
			if assert.Len(t, s1.Schedule, 1) {
				assert.EqualValues(t, "80m", s1.Schedule[0].Interval)
				assert.EqualValues(t, "", s1.Schedule[0].StartTime)
			}

			assert.Equal(t, "rbd/baz", s2.Name)
			assert.Equal(t, "4//104f1d296736", s2.LevelSpecID)
			if assert.Len(t, s2.Schedule, 1) {
				assert.EqualValues(t, "1d", s2.Schedule[0].Interval)
				assert.EqualValues(t, "14:00:00-05:00", s2.Schedule[0].StartTime)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		r := commands.NewResponse([]byte("{}"), "", nil)
		l, err := parseMirrorSnapshotScheduleList(r)
		assert.NoError(t, err)
		assert.Len(t, l, 0)
	})
	t.Run("error", func(t *testing.T) {
		r := commands.NewResponse([]byte{}, "", errors.New("yikes"))
		l, err := parseMirrorSnapshotScheduleList(r)
		assert.Error(t, err)
		assert.Len(t, l, 0)
	})
}

func TestParseMirrorSnapshotScheduleStatus(t *testing.T) {
	t.Run("status1", func(t *testing.T) {
		r := commands.NewResponse([]byte(sStatus1), "", nil)
		s, err := parseMirrorSnapshotScheduleStatus(r)
		assert.NoError(t, err)
		assert.Len(t, s, 0)
	})
	t.Run("status2", func(t *testing.T) {
		r := commands.NewResponse([]byte(sStatus2), "", nil)
		s, err := parseMirrorSnapshotScheduleStatus(r)
		assert.NoError(t, err)
		if assert.Len(t, s, 1) {
			assert.Equal(t, "rbd/foo", s[0].Image)
			assert.Contains(t, s[0].ScheduleTime, "16:30")
		}
	})
	t.Run("status3", func(t *testing.T) {
		r := commands.NewResponse([]byte(sStatus3), "", nil)
		s, err := parseMirrorSnapshotScheduleStatus(r)
		assert.NoError(t, err)
		if assert.Len(t, s, 2) {
			assert.Equal(t, "rbd/bar", s[0].Image)
			assert.Contains(t, s[0].ScheduleTime, "16:00")
			assert.Equal(t, "rbd/foo", s[1].Image)
			assert.Contains(t, s[1].ScheduleTime, "16:30")
		}
	})
	t.Run("error", func(t *testing.T) {
		r := commands.NewResponse([]byte{}, "", errors.New("zrkk"))
		s, err := parseMirrorSnapshotScheduleStatus(r)
		assert.Error(t, err)
		assert.Len(t, s, 0)
	})
}

func TestMirrorSnapshotScheduleAddRemove(t *testing.T) {
	ensureDefaultPool(t)
	ra := getAdmin(t)
	scheduler := ra.MirrorSnashotSchedule()
	t.Run("noStartTime", func(t *testing.T) {
		err := scheduler.Add(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), NoStartTime)
		assert.NoError(t, err)
		err = scheduler.Remove(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), NoStartTime)
		assert.NoError(t, err)
	})
	t.Run("startTime", func(t *testing.T) {
		stime := StartTime("12:00:00")
		err := scheduler.Add(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), stime)
		assert.NoError(t, err)
		err = scheduler.Remove(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), stime)
		assert.NoError(t, err)
	})
	t.Run("badStartTime", func(t *testing.T) {
		stime := StartTime("henry")
		err := scheduler.Add(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), stime)
		assert.Error(t, err)
	})
}

func TestMirrorSnapshotScheduleList(t *testing.T) {
	ensureDefaultPool(t)
	ra := getAdmin(t)
	// assume a pool of "rbd" exists?
	scheduler := ra.MirrorSnashotSchedule()
	err := scheduler.Add(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), NoStartTime)
	assert.NoError(t, err)
	defer func() {
		err = scheduler.Remove(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), NoStartTime)
		assert.NoError(t, err)
	}()
	slist, err := scheduler.List(NewLevelSpec(defaultPoolName, "", ""))
	assert.NoError(t, err)
	if assert.Len(t, slist, 1) {
		assert.Equal(t, "rbd/", slist[0].Name)
		if assert.Len(t, slist[0].Schedule, 1) {
			assert.Equal(t, Interval("1d"), slist[0].Schedule[0].Interval)
		}
	}

	err = scheduler.Add(NewLevelSpec(defaultPoolName, "", ""), Interval("8h"), NoStartTime)
	assert.NoError(t, err)
	defer func() {
		err = scheduler.Remove(NewLevelSpec(defaultPoolName, "", ""), Interval("8h"), NoStartTime)
		assert.NoError(t, err)
	}()
	slist, err = scheduler.List(NewLevelSpec(defaultPoolName, "", ""))
	assert.NoError(t, err)
	if assert.Len(t, slist, 1) {
		assert.Equal(t, "rbd/", slist[0].Name)
		if assert.Len(t, slist[0].Schedule, 2) {
			// ceph doesn't return the list in a "stable" order so we just
			// take the lazy approach and sort by the interval value
			sched := slist[0].Schedule
			sort.Slice(sched, func(i, j int) bool {
				return sched[i].Interval < sched[j].Interval
			})
			assert.Equal(t, Interval("1d"), sched[0].Interval)
			assert.Equal(t, Interval("8h"), sched[1].Interval)
		}
	}
}
