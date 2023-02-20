//go:build ceph_preview
// +build ceph_preview

package telemetry

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/ceph/go-ceph/internal/admintest"
)

var radosConnector = admintest.NewConnector()

type TelemetryReport struct {
	DeviceReport json.RawMessage            `json:"device_report"`
	Report       map[string]json.RawMessage `json:"report"`
}

func TestMgrAdmin_GetTelemetryStatus(t *testing.T) {
	ra := radosConnector.Get(t)
	admin := NewFromConn(ra)

	tests := []struct {
		name    string
		want    *Status
		wantErr bool
	}{
		{
			name:    "happy Path",
			want:    &Status{ChannelBasic: true, DeviceURL: "https://telemetry.ceph.com/device", Interval: 24},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := admin.Status()
			if (err != nil) != tt.wantErr {
				t.Errorf("MgrAdmin.GetTelemetryStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !(got.ChannelBasic == tt.want.ChannelBasic && got.DeviceURL == tt.want.DeviceURL && got.Interval == tt.want.Interval) {
				// if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MgrAdmin.GetTelemetryStatus() = %+v, want %v", got, tt.want)
			}
		})
	}
}

func TestMgrAdmin_GetTelemetryData(t *testing.T) {
	ra := radosConnector.Get(t)
	admin := NewFromConn(ra)

	// a map container to decode the JSON structure into
	c := TelemetryReport{
		DeviceReport: []byte(""),
		Report:       make(map[string]json.RawMessage),
	}
	tests := []struct {
		// creative name of the test
		name string
		// keys we expect in the report
		wantKeys map[string]interface{}
		// if we expect this test to error
		wantErr bool
	}{
		{
			name: "happy Path",
			wantKeys: map[string]interface{}{
				"balancer":              true,
				"channels":              true,
				"channels_available":    true,
				"collections_available": true,
				"collections_opted_in":  true,
				"config":                true,
				"crashes":               true,
				"created":               true,
				"crush":                 true,
				"fs":                    true,
				"hosts":                 true,
				"leaderboard":           true,
				"license":               true,
				"metadata":              true,
				"mon":                   true,
				"osd":                   true,
				"pools":                 true,
				"rbd":                   true,
				"report_id":             true,
				"report_timestamp":      true,
				"report_version":        true,
				"rgw":                   true,
				"rook":                  true,
				"services":              true,
				"usage":                 true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := admin.Data()
			if (err != nil) != tt.wantErr {
				t.Errorf("MgrAdmin.GetTelemetryData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			e := json.Unmarshal(got, &c)
			if (e != nil) != tt.wantErr {
				t.Errorf("Issues when unmarsheling telemetry response into JSON")
				return
			} else if (e != nil) == tt.wantErr {
				t.Logf("Error found as expected: %v", e)
				return
			}
			keys := make(map[string]interface{}, len(c.Report))
			for key := range c.Report {
				keys[strings.TrimSpace(key)] = true
			}
			if !reflect.DeepEqual(keys, tt.wantKeys) {
				t.Errorf("MgrAdmin.GetTelemetryData() = %v, want %v", keys, tt.wantKeys)
			}
		})
	}
}
