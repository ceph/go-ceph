package admin

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	authRegion        = "default"
	service           = "s3"
	connectionTimeout = time.Second * 3
	unsignedPayload   = "UNSIGNED-PAYLOAD"
)

var (
	errNoEndpoint  = errors.New("endpoint not set")
	errNoAccessKey = errors.New("access key not set")
	errNoSecretKey = errors.New("secret key not set")
)

// HTTPClient interface that conforms to that of the http package's Client.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// API struct for New Client
type API struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	HTTPClient HTTPClient
}

// New returns client for Ceph RGW
func New(endpoint, accessKey, secretKey string, httpClient HTTPClient) (*API, error) {
	// validate endpoint
	if endpoint == "" {
		return nil, errNoEndpoint
	}

	// validate access key
	if accessKey == "" {
		return nil, errNoAccessKey
	}

	// validate secret key
	if secretKey == "" {
		return nil, errNoSecretKey
	}

	// If no client is passed initialize it
	if httpClient == nil {
		httpClient = &http.Client{Timeout: connectionTimeout}
	}

	return &API{
		Endpoint:   endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		HTTPClient: httpClient,
	}, nil
}

// doRequest signs and sends an HTTP request, then reads and returns the
// response body. The payloadHash is included in the S3 v4 signature;
// use "UNSIGNED-PAYLOAD" when the body hash is not required by the server.
func (api *API) doRequest(ctx context.Context, req *http.Request, payloadHash string) ([]byte, error) {
	credCache := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(api.AccessKey, api.SecretKey, ""))
	creds, err := credCache.Retrieve(ctx)
	if err != nil {
		return nil, err
	}

	signer := v4.NewSigner()
	err = signer.SignHTTP(ctx, creds, req, payloadHash, service, authRegion, time.Now())
	if err != nil {
		return nil, err
	}

	resp, err := api.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	if resp.StatusCode >= 300 {
		return nil, handleStatusError(body)
	}

	return body, nil
}

// call makes request to the RGW Admin Ops API
func (api *API) call(ctx context.Context, httpMethod, path string, args url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, httpMethod, buildQueryPath(api.Endpoint, path, args.Encode()), nil)
	if err != nil {
		return nil, err
	}

	return api.doRequest(ctx, req, unsignedPayload)
}

// sha256Hex returns the hex-encoded SHA-256 digest of data.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// callSNS makes a POST request to the SNS-compatible topic API endpoint.
// The SNS endpoint is at the S3 root, and is dispatched by the "Action" parameter.
// All parameters (Action, Name, TopicArn, Attributes) are sent in the POST form
// body because RGW's parse_post_action() only recognizes the
// Attributes.entry.N.{key|value} format from the request body.
func (api *API) callSNS(ctx context.Context, action string, params url.Values) ([]byte, error) {
	if params == nil {
		params = url.Values{}
	}
	params.Set("Action", action)

	body := []byte(params.Encode())
	payloadHash := sha256Hex(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, api.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return api.doRequest(ctx, req, payloadHash)
}
