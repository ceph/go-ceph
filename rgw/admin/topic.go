//go:build ceph_preview

package admin

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"sort"
)

// Topic represents an SNS-compatible topic
type Topic struct {
	User       string        `xml:"User"`
	Name       string        `xml:"Name"`
	EndPoint   TopicEndPoint `xml:"EndPoint"`
	TopicArn   string        `xml:"TopicArn"`
	OpaqueData string        `xml:"OpaqueData"`
	Policy     string        `xml:"Policy"`
}

// TopicEndPoint represents the endpoint configuration for a topic
type TopicEndPoint struct {
	EndpointAddress    string     `xml:"EndpointAddress" json:"EndpointAddress"`
	EndpointArgs       string     `xml:"EndpointArgs" json:"EndpointArgs"`
	EndpointTopic      string     `xml:"EndpointTopic" json:"EndpointTopic"`
	HasStoredSecret    BoolString `xml:"HasStoredSecret" json:"HasStoredSecret"`
	Persistent         BoolString `xml:"Persistent" json:"Persistent"`
	TimeToLive         string     `xml:"TimeToLive" json:"TimeToLive"`
	MaxRetries         string     `xml:"MaxRetries" json:"MaxRetries"`
	RetrySleepDuration string     `xml:"RetrySleepDuration" json:"RetrySleepDuration"`
}

// BoolString unmarshals JSON values that may be either a string ("true"/"false")
// or a boolean (true/false), for cross-version Ceph compatibility.
type BoolString bool

// UnmarshalJSON supports unmarshalling a BoolString from JSON.
func (bs *BoolString) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*bs = s == "true" || s == "True"
		return nil
	}
	var b bool
	if err := json.Unmarshal(data, &b); err == nil {
		*bs = BoolString(b)
		return nil
	}
	return fmt.Errorf("cannot unmarshal %s into BoolString", string(data))
}

// CreateTopicResponse represents the response from a CreateTopic request
type CreateTopicResponse struct {
	XMLName           xml.Name `xml:"CreateTopicResponse"`
	CreateTopicResult struct {
		TopicArn string `xml:"TopicArn"`
	} `xml:"CreateTopicResult"`
}

// GetTopicAttributesResponse represents the response from a GetTopicAttributes request
type GetTopicAttributesResponse struct {
	XMLName                  xml.Name `xml:"GetTopicAttributesResponse"`
	GetTopicAttributesResult struct {
		Attributes struct {
			Entry []TopicAttribute `xml:"entry"`
		} `xml:"Attributes"`
	} `xml:"GetTopicAttributesResult"`
}

// TopicAttribute represents a single key-value attribute of a topic
type TopicAttribute struct {
	Key   string `xml:"key"`
	Value string `xml:"value"`
}

// ListTopicsResponse represents the response from a ListTopics request
type ListTopicsResponse struct {
	XMLName          xml.Name `xml:"ListTopicsResponse"`
	ListTopicsResult struct {
		Topics struct {
			Topic []Topic `xml:"member"`
		} `xml:"Topics"`
	} `xml:"ListTopicsResult"`
}

// CreateTopic creates a new SNS-compatible topic
// https://docs.ceph.com/en/latest/radosgw/s3/bucketops/#create-topic
func (api *API) CreateTopic(ctx context.Context, name string, attrs map[string]string) (string, error) {
	params := url.Values{}
	params.Set("Name", name)

	keys := make([]string, 0, len(attrs))
	for k := range attrs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		params.Set(fmt.Sprintf("Attributes.entry.%d.key", i+1), k)
		params.Set(fmt.Sprintf("Attributes.entry.%d.value", i+1), attrs[k])
	}

	body, err := api.callSNS(ctx, "CreateTopic", params)
	if err != nil {
		return "", err
	}

	var resp CreateTopicResponse
	err = xml.Unmarshal(body, &resp)
	if err != nil {
		return "", fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return resp.CreateTopicResult.TopicArn, nil
}

// GetTopicAttributes returns information about a specific topic
func (api *API) GetTopicAttributes(ctx context.Context, topicARN string) (Topic, error) {
	params := url.Values{}
	params.Set("TopicArn", topicARN)

	body, err := api.callSNS(ctx, "GetTopicAttributes", params)
	if err != nil {
		return Topic{}, err
	}

	var resp GetTopicAttributesResponse
	err = xml.Unmarshal(body, &resp)
	if err != nil {
		return Topic{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	var topic Topic
	for _, entry := range resp.GetTopicAttributesResult.Attributes.Entry {
		switch entry.Key {
		case "User":
			topic.User = entry.Value
		case "Name":
			topic.Name = entry.Value
		case "TopicArn":
			topic.TopicArn = entry.Value
		case "OpaqueData":
			topic.OpaqueData = entry.Value
		case "Policy":
			topic.Policy = entry.Value
		case "EndPoint":
			if err := json.Unmarshal([]byte(entry.Value), &topic.EndPoint); err != nil {
				return Topic{}, fmt.Errorf("failed to parse EndPoint JSON: %w", err)
			}
		}
	}

	return topic, nil
}

// DeleteTopic deletes a topic
func (api *API) DeleteTopic(ctx context.Context, topicARN string) error {
	params := url.Values{}
	params.Set("TopicArn", topicARN)

	_, err := api.callSNS(ctx, "DeleteTopic", params)
	return err
}

// ListTopics lists all topics for the tenant
func (api *API) ListTopics(ctx context.Context) ([]Topic, error) {
	body, err := api.callSNS(ctx, "ListTopics", nil)
	if err != nil {
		return nil, err
	}

	var resp ListTopicsResponse
	err = xml.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return resp.ListTopicsResult.Topics.Topic, nil
}
