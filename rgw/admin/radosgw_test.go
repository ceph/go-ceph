package admin

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RadosGWTestSuite struct {
	suite.Suite
	endpoint  string
	accessKey string
	secretKey string
}

func TestRadosGWTestSuite(t *testing.T) {
	suite.Run(t, new(RadosGWTestSuite))
}

func (suite *RadosGWTestSuite) SetupConnection() {
	suite.accessKey = "2262XNX11FZRR44XWIRD"
	suite.secretKey = "rmtuS1Uj1bIC08QFYGW18GfSHAbkPqdsuYynNudw"
	hostname := os.Getenv("HOSTNAME")
	endpoint := hostname
	if hostname != "test_ceph_aio" {
		endpoint = "test_ceph_a"
	}
	suite.endpoint = "http://" + endpoint
}

func TestNew(t *testing.T) {
	type args struct {
		endpoint  string
		accessKey string
		secretKey string
	}
	tests := []struct {
		name    string
		args    args
		want    *API
		wantErr error
	}{
		{"no endpoint", args{}, nil, errNoEndpoint},
		{"no accessKey", args{endpoint: "http://192.168.0.1"}, nil, errNoAccessKey},
		{"no secretKey", args{endpoint: "http://192.168.0.1", accessKey: "foo"}, nil, errNoSecretKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.endpoint, tt.args.accessKey, tt.args.secretKey, nil)
			if tt.wantErr != err {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
