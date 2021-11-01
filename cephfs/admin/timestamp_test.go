package admin

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeStampUnmarshal(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		j1 := []byte(`"2020-01-03 18:03:21"`)
		var ts TimeStamp
		err := json.Unmarshal(j1, &ts)
		assert.NoError(t, err)
		assert.Equal(t, 2020, ts.Year())
		assert.Equal(t, time.Month(1), ts.Month())
		assert.Equal(t, 3, ts.Day())
	})
	t.Run("badType", func(t *testing.T) {
		j1 := []byte(`["2020-01-03 18:03:21"]`)
		var ts TimeStamp
		err := json.Unmarshal(j1, &ts)
		assert.Error(t, err)
	})
	t.Run("badValue", func(t *testing.T) {
		j1 := []byte(`"just another manic monday"`)
		var ts TimeStamp
		err := json.Unmarshal(j1, &ts)
		assert.Error(t, err)
	})
}

func TestTimeStampString(t *testing.T) {
	s := "2020-11-06 11:33:56"
	ti, err := time.Parse(cephTSLayout, s)
	if assert.NoError(t, err) {
		ts := TimeStamp{ti}
		assert.Equal(t, s, ts.String())
	}
}
