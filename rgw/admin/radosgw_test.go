package admin

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/suite"
)

type RadosGWTestSuite struct {
	suite.Suite
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
	Client *s3.S3
}

func newS3Agent(accessKey, secretKey, endpoint string, debug bool) (*S3Agent, error) {
	const cephRegion = "us-east-1"

	logLevel := aws.LogOff
	if debug {
		logLevel = aws.LogDebug
	}
	client := http.Client{
		Timeout: time.Second * 15,
	}
	sess, err := session.NewSession(
		aws.NewConfig().
			WithRegion(cephRegion).
			WithCredentials(credentials.NewStaticCredentials(accessKey, secretKey, "")).
			WithEndpoint(endpoint).
			WithS3ForcePathStyle(true).
			WithMaxRetries(5).
			WithDisableSSL(true).
			WithHTTPClient(&client).
			WithLogLevel(logLevel),
	)
	if err != nil {
		return nil, err
	}
	svc := s3.New(sess)
	return &S3Agent{
		Client: svc,
	}, nil
}

func (s *S3Agent) createBucket(name string) error {
	bucketInput := &s3.CreateBucketInput{
		Bucket: &name,
	}
	_, err := s.Client.CreateBucket(bucketInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeBucketAlreadyExists:
				return nil
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				return nil
			}
		}
		return fmt.Errorf("failed to create bucket %q. %w", name, err)
	}
	return nil
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
