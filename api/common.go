package api

import (
	"fmt"
	"net/http"
	"io/ioutil"
)

type CephClient struct {
	BaseUrl string // e.g. http://<ceph-rest-api>:5000/v1/api/
}

func (cc *CephClient) callApi(endpoint string, method string) (string, error) {
	var body string
	endpoint = cc.BaseUrl + endpoint
	client := http.Client{}

	fmt.Printf("METHOD: %s\n", method)
	fmt.Printf("Endpoint: %s\n", endpoint)

	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		return "", err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return body, err
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}
	return string(bodyBytes), nil
}
