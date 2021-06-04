package admin

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"errors"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
)

const (
	authRegion                 = "default"
	service                    = "s3"
	connectionTimeout          = time.Second * 3
	post              verbHTTP = "POST"
	put               verbHTTP = "PUT"
	get               verbHTTP = "GET"
	delete            verbHTTP = "DELETE"
)

var (
	errNoEndpoint  = errors.New("endpoint not set")
	errNoAccessKey = errors.New("access key not set")
	errNoSecretKey = errors.New("secret key not set")
)

type verbHTTP string

// API struct for New Client
type API struct {
	AccessKey  string
	SecretKey  string
	Endpoint   string
	HTTPClient *http.Client
	Debug      bool
}

// New returns client for Ceph RGW
func New(endpoint, accessKey, secretKey string, httpClient *http.Client) (*API, error) {
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

	// set http client, if TLS is desired it will have to be passed with an http client
	if httpClient == nil {
		httpClient = &http.Client{Timeout: connectionTimeout}
	}

	return &API{
		Endpoint:   endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		HTTPClient: httpClient,
		Debug:      false,
	}, nil
}

// call makes request to the RGW Admin Ops API
func (api *API) call(ctx context.Context, verb verbHTTP, path string, args url.Values) (body []byte, err error) {
	// Build request
	request, err := http.NewRequestWithContext(ctx, string(verb), buildQueryPath(api.Endpoint, path, args.Encode()), nil)
	if err != nil {
		return nil, err
	}

	// Build S3 authentication
	cred := credentials.NewStaticCredentials(api.AccessKey, api.SecretKey, "")
	signer := v4.NewSigner(cred)
	// This was present in https://github.com/IrekFasikhov/go-rgwadmin/ but it seems that the lib works without it
	// Let's keep it here just in case something shows up
	// signer.DisableRequestBodyOverwrite = true

	// Sign in S3
	_, err = signer.Sign(request, nil, service, authRegion, time.Now())
	if err != nil {
		return nil, err
	}

	// Print request if Debug is enabled
	if api.Debug {
		dump, err := httputil.DumpRequestOut(request, true)
		if err != nil {
			return nil, err
		}
		log.Printf("\n%s\n", string(dump))
	}

	// Send HTTP request
	resp, err := api.HTTPClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Print response if Debug is enabled
	if api.Debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			return nil, err
		}
		log.Printf("\n%s\n", string(dump))
	}

	// Decode HTTP response
	decodedResponse, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(decodedResponse))

	// Handle error in response
	if resp.StatusCode >= 300 {
		return nil, handleStatusError(decodedResponse)
	}

	return decodedResponse, nil
}
