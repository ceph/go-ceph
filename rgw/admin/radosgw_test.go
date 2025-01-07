package admin

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/logging"
	tsuite "github.com/stretchr/testify/suite"
)

type RadosGWTestSuite struct {
	tsuite.Suite
	endpoint       string
	accessKey      string
	secretKey      string
	bucketTestName string
}

type debugHTTPClient struct {
	client HTTPClient
}

func newDebugHTTPClient(client HTTPClient) *debugHTTPClient {
	return &debugHTTPClient{client}
}

func (c *debugHTTPClient) Do(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n%s\n", string(dump))

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	fmt.Printf("\n%s\n", string(dump))

	return resp, nil
}

func TestRadosGWTestSuite(t *testing.T) {
	tsuite.Run(t, new(RadosGWTestSuite))
}

// S3Agent wraps the s3.S3 structure to allow for wrapper methods
type S3Agent struct {
	Client *s3.Client
}

func newS3Agent(accessKey, secretKey, endpoint string, debug bool) (*S3Agent, error) {
	const cephRegion = "us-east-1"

	var logger logging.Logger
	logger = logging.Nop{}
	if debug {
		logger = logging.NewStandardLogger(os.Stderr)
	}

	client := http.Client{
		Timeout: time.Second * 15,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cephRegion),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     accessKey,
				SecretAccessKey: secretKey,
				SessionToken:    "",
			},
		}),
		config.WithBaseEndpoint(endpoint),
		config.WithRetryMaxAttempts(5),
		config.WithHTTPClient(&client),
		config.WithLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	return &S3Agent{
		Client: svc,
	}, nil
}

func (s *S3Agent) createBucket(name string) error {
	bucketInput := &s3.CreateBucketInput{
		Bucket: &name,
	}

	var (
		bae   *types.BucketAlreadyExists
		baoby *types.BucketAlreadyOwnedByYou
	)

	_, err := s.Client.CreateBucket(context.TODO(), bucketInput)
	switch {
	case err == nil:
	case errors.As(err, &bae):
	case errors.As(err, &baoby):
	default:
		return fmt.Errorf("failed to create bucket %q. %w", name, err)
	}

	return nil
}

func (suite *RadosGWTestSuite) SetupConnection() {
	suite.accessKey = "AKIAIOSFODNN7EXAMPLE"
	suite.secretKey = "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
	hostname := os.Getenv("HOSTNAME")
	endpoint := hostname
	if hostname != "test_ceph_aio" {
		endpoint = "test_ceph_a"
	}
	suite.endpoint = "http://" + endpoint
	suite.bucketTestName = "test"
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
