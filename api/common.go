package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type CephClient struct {
	BaseUrl string // e.g. http://<ceph-rest-api>:5000/v1/api/
}

func (cc *CephClient) callApi(endpoint string, method string) (string, error) {
	var body string
	endpoint = cc.BaseUrl + endpoint
	client := http.Client{
		Timeout: 5 * time.Minute,
	}
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	log.Printf("Sending request to ceph-rest-api with endpoint %s", endpoint)
	resp, err := client.Do(req)
	log.Printf("Got request response to ceph-rest-api with endpoint %s", endpoint)
	if err != nil {
		return body, err
	}

	if resp.StatusCode != http.StatusOK {
		return body, fmt.Errorf("Received unexpected status code from server: %d", resp.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return string(bodyBytes), nil
}
