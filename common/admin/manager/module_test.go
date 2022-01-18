package manager

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ceph/go-ceph/internal/admintest"
	"github.com/ceph/go-ceph/internal/commands"
)

var radosConnector = admintest.NewConnector()

func TestParseModuleInfo(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		r := commands.NewResponse([]byte(cephMgrModuleLsJSON1), "", nil)
		mi, err := parseModuleInfo(r)
		assert.NoError(t, err)
		if assert.Len(t, mi.EnabledModules, 7) {
			assert.Contains(t, mi.EnabledModules, "mirroring")
			assert.Contains(t, mi.EnabledModules, "nfs")
			assert.NotContains(t, mi.EnabledModules, "progress")
		}
		if assert.Len(t, mi.AlwaysOnModules, 10) {
			assert.Contains(t, mi.AlwaysOnModules, "balancer")
			assert.Contains(t, mi.AlwaysOnModules, "volumes")
			assert.NotContains(t, mi.AlwaysOnModules, "notebook")
		}
		if assert.Len(t, mi.DisabledModules, 19) {
			assert.Equal(t, mi.DisabledModules[0].Name, "alerts")
			assert.True(t, mi.DisabledModules[0].CanRun)

			assert.Contains(t, mi.DisabledModules[4].Name, "influx")
			assert.False(t, mi.DisabledModules[4].CanRun)

			assert.Contains(t, mi.DisabledModules[10].Name, "osd_support")
			assert.True(t, mi.DisabledModules[10].CanRun)
		}
	})
	t.Run("error", func(t *testing.T) {
		r := commands.NewResponse(nil, "", errors.New("foo"))
		mi, err := parseModuleInfo(r)
		assert.Error(t, err)
		assert.Nil(t, mi)
	})
}

func TestEnableDisableModule(t *testing.T) {
	ra := radosConnector.Get(t)
	mgrAdmin := NewFromConn(ra)

	mi, err := mgrAdmin.ListModules()
	require.NoError(t, err)

	var toEnable string
dmloop:
	for _, entry := range mi.DisabledModules {
		if !entry.CanRun {
			continue
		}
		switch entry.Name {
		// not all modules in the disabled list work (in our container?).
		// here's a short list our test is allowed to try enabling
		case "hello", "mirroring", "rgw", "prometheus", "dashboard":
			toEnable = entry.Name
			break dmloop
		default:
		}
	}
	require.NotEqual(t, "", toEnable)

	err = mgrAdmin.EnableModule(toEnable, false)
	require.NoError(t, err)
	waitFor(t, mgrAdmin, toEnable)

	// put it back to disabled state
	err = mgrAdmin.DisableModule(toEnable)
	require.NoError(t, err)
}

func waitFor(t *testing.T, mgrAdmin *MgrAdmin, expect string) {
	for i := 0; i < 30; i++ {
		modinfo, err := mgrAdmin.ListModules()
		require.NoError(t, err)
		for _, emod := range modinfo.EnabledModules {
			if emod == expect {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s module", expect)
}

var cephMgrModuleLsJSON1 = `
{
    "always_on_modules": [
        "balancer",
        "crash",
        "devicehealth",
        "orchestrator",
        "pg_autoscaler",
        "progress",
        "rbd_support",
        "status",
        "telemetry",
        "volumes"
    ],
    "enabled_modules": [
        "cephadm",
        "dashboard",
        "iostat",
        "mirroring",
        "nfs",
        "prometheus",
        "restful"
    ],
    "disabled_modules": [
        {
            "name": "alerts",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "interval": {
                    "name": "interval",
                    "type": "secs",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "60",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "How frequently to reexamine health status",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_destination": {
                    "name": "smtp_destination",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Email address to send alerts to",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_from_name": {
                    "name": "smtp_from_name",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "Ceph",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Email From: name",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_host": {
                    "name": "smtp_host",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "SMTP server",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_password": {
                    "name": "smtp_password",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Password to authenticate with",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_port": {
                    "name": "smtp_port",
                    "type": "int",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "465",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "SMTP port",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_sender": {
                    "name": "smtp_sender",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "SMTP envelope sender",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_ssl": {
                    "name": "smtp_ssl",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "True",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Use SSL to connect to SMTP server",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "smtp_user": {
                    "name": "smtp_user",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "User to authenticate as",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "cli_api",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "diskprediction_local",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "predict_interval": {
                    "name": "predict_interval",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "86400",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "predictor_model": {
                    "name": "predictor_model",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "prophetstor",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "sleep_interval": {
                    "name": "sleep_interval",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "600",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "hello",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "emphatic": {
                    "name": "emphatic",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "True",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "whether to say it loudly",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "foo": {
                    "name": "foo",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "a",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "a",
                        "b",
                        "c"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "place": {
                    "name": "place",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "world",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "a place in the world",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "influx",
            "can_run": false,
            "error_string": "influxdb python module not found",
            "module_options": {
                "batch_size": {
                    "name": "batch_size",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "5000",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "How big batches of data points should be when sending to InfluxDB.",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "database": {
                    "name": "database",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "ceph",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "InfluxDB database name. You will need to create this database and grant write privileges to the configured username or the username must have admin privileges to create it.",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "hostname": {
                    "name": "hostname",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "InfluxDB server hostname",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "interval": {
                    "name": "interval",
                    "type": "secs",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "30",
                    "min": "5",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Time between reports to InfluxDB.  Default 30 seconds.",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "password": {
                    "name": "password",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "password of InfluxDB server user",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "port": {
                    "name": "port",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "8086",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "InfluxDB server port",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "ssl": {
                    "name": "ssl",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "false",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Use https connection for InfluxDB server. Use \"true\" or \"false\".",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "threads": {
                    "name": "threads",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "5",
                    "min": "1",
                    "max": "32",
                    "enum_allowed": [],
                    "desc": "How many worker threads should be spawned for sending data to InfluxDB.",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "username": {
                    "name": "username",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "username of InfluxDB server user",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "verify_ssl": {
                    "name": "verify_ssl",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "true",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Verify https cert for InfluxDB server. Use \"true\" or \"false\".",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "insights",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "k8sevents",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "ceph_event_retention_days": {
                    "name": "ceph_event_retention_days",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "7",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "Days to hold ceph event information within local cache",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "config_check_secs": {
                    "name": "config_check_secs",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "10",
                    "min": "10",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "interval (secs) to check for cluster configuration changes",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "localpool",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "failure_domain": {
                    "name": "failure_domain",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "host",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "failure domain for any created local pool",
                    "long_desc": "what failure domain we should separate data replicas across.",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "min_size": {
                    "name": "min_size",
                    "type": "int",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "default min_size for any created local pool",
                    "long_desc": "value to set min_size to (unchanged from Ceph's default if this option is not set)",
                    "tags": [],
                    "see_also": []
                },
                "num_rep": {
                    "name": "num_rep",
                    "type": "int",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "3",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "default replica count for any created local pool",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "pg_num": {
                    "name": "pg_num",
                    "type": "int",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "128",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "default pg_num for any created local pool",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "prefix": {
                    "name": "prefix",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "name prefix for any created local pool",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "subtree": {
                    "name": "subtree",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "rack",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "CRUSH level for which to create a local pool",
                    "long_desc": "which CRUSH subtree type the module should create a pool for.",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "mds_autoscaler",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "osd_perf_query",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "osd_support",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "rgw",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "rook",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "drive_group_interval": {
                    "name": "drive_group_interval",
                    "type": "float",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "300.0",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "interval in seconds between re-application of applied drive_groups",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "storage_class": {
                    "name": "storage_class",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "local",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "storage class name for LSO-discovered PVs",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "selftest",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "roption1": {
                    "name": "roption1",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "roption2": {
                    "name": "roption2",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "xyz",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "rwoption1": {
                    "name": "rwoption1",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "rwoption2": {
                    "name": "rwoption2",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "rwoption3": {
                    "name": "rwoption3",
                    "type": "float",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "rwoption4": {
                    "name": "rwoption4",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "rwoption5": {
                    "name": "rwoption5",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "rwoption6": {
                    "name": "rwoption6",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "True",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "rwoption7": {
                    "name": "rwoption7",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "1",
                    "max": "42",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "testkey": {
                    "name": "testkey",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "testlkey": {
                    "name": "testlkey",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "testnewline": {
                    "name": "testnewline",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "snap_schedule",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "allow_m_granularity": {
                    "name": "allow_m_granularity",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "allow minute scheduled snapshots",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "dump_on_update": {
                    "name": "dump_on_update",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "dump database to debug log on update",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "stats",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "telegraf",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "address": {
                    "name": "address",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "unixgram:///tmp/telegraf.sock",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "interval": {
                    "name": "interval",
                    "type": "secs",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "15",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "test_orchestrator",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        },
        {
            "name": "zabbix",
            "can_run": true,
            "error_string": "",
            "module_options": {
                "discovery_interval": {
                    "name": "discovery_interval",
                    "type": "uint",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "100",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "identifier": {
                    "name": "identifier",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "interval": {
                    "name": "interval",
                    "type": "secs",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "60",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_level": {
                    "name": "log_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster": {
                    "name": "log_to_cluster",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_cluster_level": {
                    "name": "log_to_cluster_level",
                    "type": "str",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "info",
                    "min": "",
                    "max": "",
                    "enum_allowed": [
                        "",
                        "critical",
                        "debug",
                        "error",
                        "info",
                        "warning"
                    ],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "log_to_file": {
                    "name": "log_to_file",
                    "type": "bool",
                    "level": "advanced",
                    "flags": 1,
                    "default_value": "False",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "zabbix_host": {
                    "name": "zabbix_host",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "zabbix_port": {
                    "name": "zabbix_port",
                    "type": "int",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "10051",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                },
                "zabbix_sender": {
                    "name": "zabbix_sender",
                    "type": "str",
                    "level": "advanced",
                    "flags": 0,
                    "default_value": "/usr/bin/zabbix_sender",
                    "min": "",
                    "max": "",
                    "enum_allowed": [],
                    "desc": "",
                    "long_desc": "",
                    "tags": [],
                    "see_also": []
                }
            }
        }
    ]
}`
