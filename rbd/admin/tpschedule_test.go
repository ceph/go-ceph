//go:build ceph_preview

package admin

import (
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ceph/go-ceph/internal/commands"
)

var tpsList1 = `
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
    "4/myns": {
        "name": "mypool/myns/",
        "schedule": [
            {
                "interval": "99m",
                "start_time": null
            }
        ]
    }
}
`

var tpsList2 = `
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
    "4/myns": {
        "name": "mypool/myns/",
        "schedule": [
            {
                "interval": "1d",
                "start_time": "2021-08-02 14:00:00"
            }
        ]
    }
}
`

var tpsStatus1 = `
{
    "scheduled": []
}
`
var tpsStatus2 = `
{
    "scheduled": [
        {
            "pool_name": "rbd",
            "pool_id": "4",
            "namespace": "",
            "schedule_time": "2021-08-02 11:50:00"
        }
    ]
}
`

var tpsStatus3 = `
{
    "scheduled": [
        {
            "pool_name": "mypool",
            "pool_id": "5",
            "namespace": "ns1",
            "schedule_time": "2021-08-02 11:00:00"
        },
        {
            "pool_name": "rbd",
            "pool_id": "4",
            "namespace": "",
            "schedule_time": "2021-08-02 11:50:00"
        }
    ]
}
`

func TestParseTrashPurgeScheduleList(t *testing.T) {
	t.Run("list1", func(t *testing.T) {
		r := commands.NewResponse([]byte(tpsList1), "", nil)
		l, err := parseTrashPurgeScheduleList(r)
		assert.NoError(t, err)
		if assert.Len(t, l, 2) {
			s1 := l[0]
			s2 := l[1]
			if s1.Name != "rbd/" {
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

			assert.Equal(t, "mypool/myns/", s2.Name)
			assert.Equal(t, "4/myns", s2.LevelSpecID)
			if assert.Len(t, s2.Schedule, 1) {
				assert.EqualValues(t, "99m", s2.Schedule[0].Interval)
				assert.EqualValues(t, "", s2.Schedule[0].StartTime)
			}
		}
	})
	t.Run("list2", func(t *testing.T) {
		r := commands.NewResponse([]byte(tpsList2), "", nil)
		l, err := parseTrashPurgeScheduleList(r)
		assert.NoError(t, err)
		if assert.Len(t, l, 2) {
			s1 := l[0]
			s2 := l[1]
			if s1.Name != "rbd/" {
				s1, s2 = s2, s1
			}
			assert.Equal(t, "rbd/", s1.Name)
			assert.Equal(t, "4", s1.LevelSpecID)
			if assert.Len(t, s1.Schedule, 1) {
				assert.EqualValues(t, "80m", s1.Schedule[0].Interval)
				assert.EqualValues(t, "", s1.Schedule[0].StartTime)
			}

			assert.Equal(t, "mypool/myns/", s2.Name)
			assert.Equal(t, "4/myns", s2.LevelSpecID)
			if assert.Len(t, s2.Schedule, 1) {
				assert.EqualValues(t, "1d", s2.Schedule[0].Interval)
				assert.EqualValues(t, "2021-08-02 14:00:00", s2.Schedule[0].StartTime)
			}
		}
	})
	t.Run("empty", func(t *testing.T) {
		r := commands.NewResponse([]byte("{}"), "", nil)
		l, err := parseTrashPurgeScheduleList(r)
		assert.NoError(t, err)
		assert.Len(t, l, 0)
	})
	t.Run("error", func(t *testing.T) {
		r := commands.NewResponse([]byte{}, "", errors.New("yikes"))
		l, err := parseTrashPurgeScheduleList(r)
		assert.Error(t, err)
		assert.Len(t, l, 0)
	})
}

func TestParseTrashPurgeScheduleStatus(t *testing.T) {
	t.Run("status1", func(t *testing.T) {
		r := commands.NewResponse([]byte(tpsStatus1), "", nil)
		s, err := parseTrashPurgeScheduleStatus(r)
		assert.NoError(t, err)
		assert.Len(t, s, 0)
	})
	t.Run("status2", func(t *testing.T) {
		r := commands.NewResponse([]byte(tpsStatus2), "", nil)
		s, err := parseTrashPurgeScheduleStatus(r)
		assert.NoError(t, err)
		if assert.Len(t, s, 1) {
			assert.Equal(t, "rbd", s[0].PoolName)
			assert.Equal(t, "4", s[0].PoolID)
			assert.Equal(t, "", s[0].Namespace)
			assert.Contains(t, s[0].ScheduleTime, "11:50")
		}
	})
	t.Run("status3", func(t *testing.T) {
		r := commands.NewResponse([]byte(tpsStatus3), "", nil)
		s, err := parseTrashPurgeScheduleStatus(r)
		assert.NoError(t, err)
		if assert.Len(t, s, 2) {
			assert.Equal(t, "mypool", s[0].PoolName)
			assert.Equal(t, "5", s[0].PoolID)
			assert.Equal(t, "ns1", s[0].Namespace)
			assert.Contains(t, s[0].ScheduleTime, "11:00")
			assert.Equal(t, "rbd", s[1].PoolName)
			assert.Equal(t, "4", s[1].PoolID)
			assert.Equal(t, "", s[1].Namespace)
			assert.Contains(t, s[1].ScheduleTime, "11:50")
		}
	})
	t.Run("error", func(t *testing.T) {
		r := commands.NewResponse([]byte{}, "", errors.New("zrkk"))
		s, err := parseTrashPurgeScheduleStatus(r)
		assert.Error(t, err)
		assert.Len(t, s, 0)
	})
}

func TestTrashPurgeScheduleAddRemove(t *testing.T) {
	ensureDefaultPool(t)
	ra := getAdmin(t)
	scheduler := ra.TrashPurgeSchedule()
	t.Run("noStartTime", func(t *testing.T) {
		err := scheduler.Add(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), NoStartTime)
		assert.NoError(t, err)
		err = scheduler.Remove(NewLevelSpec(defaultPoolName, "", ""), Interval("1d"), NoStartTime)
		assert.NoError(t, err)
	})
	t.Run("startTime", func(t *testing.T) {
		stime := StartTime(time.Now().Format("2006-01-02T15:04:00"))
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

func TestTrashPurgeScheduleList(t *testing.T) {
	ensureDefaultPool(t)
	ra := getAdmin(t)
	scheduler := ra.TrashPurgeSchedule()
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
			sched := slist[0].Schedule
			sort.Slice(sched, func(i, j int) bool {
				return sched[i].Interval < sched[j].Interval
			})
			assert.Equal(t, Interval("1d"), sched[0].Interval)
			assert.Equal(t, Interval("8h"), sched[1].Interval)
		}
	}
}
