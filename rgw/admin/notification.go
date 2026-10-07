//go:build ceph_preview

package admin

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
)

// NotificationConfiguration represents the S3 notification configuration
type NotificationConfiguration struct {
	XMLName             xml.Name             `xml:"NotificationConfiguration"`
	TopicConfigurations []TopicConfiguration `xml:"TopicConfiguration"`
}

// TopicConfiguration represents a topic configuration for notifications
type TopicConfiguration struct {
	Id     string   `xml:"Id"`
	Topic  string   `xml:"Topic"`
	Event  []string `xml:"Event"`
	Filter *Filter  `xml:"Filter,omitempty"`
}

// Filter represents the filter for notification events
type Filter struct {
	S3Key      *FilterContainer `xml:"S3Key,omitempty"`
	S3Metadata *FilterContainer `xml:"S3Metadata,omitempty"`
	S3Tags     *FilterContainer `xml:"S3Tags,omitempty"`
}

// FilterContainer holds a list of filter rules
type FilterContainer struct {
	FilterRules []FilterRule `xml:"FilterRule"`
}

// FilterRule represents a single filter rule
type FilterRule struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

// CreateNotification creates a notification configuration for a bucket
// https://docs.ceph.com/en/latest/radosgw/s3/bucketops/#create-notification
func (api *API) CreateNotification(ctx context.Context, bucket Bucket, config NotificationConfiguration) error {
	if bucket.Bucket == "" {
		return errMissingBucket
	}

	xmlBody, err := xml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal notification configuration: %w", err)
	}

	xmlBody = append([]byte(xml.Header), xmlBody...)

	_, err = api.callNotification(ctx, http.MethodPut, bucket.Bucket, "", xmlBody)
	return err
}

// DeleteNotification deletes a specific notification or all notifications from a bucket
// If notificationID is empty, all notifications are deleted
// https://docs.ceph.com/en/latest/radosgw/s3/bucketops/#delete-notification
func (api *API) DeleteNotification(ctx context.Context, bucket Bucket, notificationID string) error {
	if bucket.Bucket == "" {
		return errMissingBucket
	}

	_, err := api.callNotification(ctx, http.MethodDelete, bucket.Bucket, notificationID, nil)
	return err
}

// GetNotifications retrieves notification configurations for a bucket
// If notificationID is empty, all notifications are returned
// https://docs.ceph.com/en/latest/radosgw/s3/bucketops/#get-list-notification
func (api *API) GetNotifications(ctx context.Context, bucket Bucket, notificationID string) (NotificationConfiguration, error) {
	if bucket.Bucket == "" {
		return NotificationConfiguration{}, errMissingBucket
	}

	body, err := api.callNotification(ctx, http.MethodGet, bucket.Bucket, notificationID, nil)
	if err != nil {
		return NotificationConfiguration{}, err
	}

	var config NotificationConfiguration
	err = xml.Unmarshal(body, &config)
	if err != nil {
		return NotificationConfiguration{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return config, nil
}
