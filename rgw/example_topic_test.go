//go:build ceph_preview

package rgw_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

// Example_topics demonstrates RGW topic management using the official SNS SDK.
func Example_topics() {
	ctx := context.Background()
	client := sns.New(sns.Options{
		BaseEndpoint: aws.String(os.Getenv("RGW_ENDPOINT")),
		Region:       "default",
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("RGW_ACCESS_KEY"), os.Getenv("RGW_SECRET_KEY"), ""),
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	})

	// Ceph-specific attributes fit the SDK's existing Attributes map.
	topic, err := client.CreateTopic(ctx, &sns.CreateTopicInput{
		Name: aws.String(fmt.Sprintf("go-ceph-example-%d", time.Now().UnixNano())),
		Attributes: map[string]string{
			"push-endpoint": "http://localhost:8080",
			"persistent":    "true",
			"OpaqueData":    "tenant=acme;pipeline=ingest",
		},
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		_, err := client.DeleteTopic(ctx, &sns.DeleteTopicInput{TopicArn: topic.TopicArn})
		if err != nil {
			panic(err)
		}
	}()

	attributes, err := client.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
		TopicArn: topic.TopicArn,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("Topic ARN:", attributes.Attributes["TopicArn"])
	fmt.Println("Opaque data:", attributes.Attributes["OpaqueData"])

	// RGW returns EndPoint as JSON inside the XML attribute value. Quincy uses
	// strings for boolean fields; newer versions use JSON booleans.
	var endpoint map[string]interface{}
	if err := json.Unmarshal([]byte(attributes.Attributes["EndPoint"]), &endpoint); err != nil {
		panic(err)
	}
	fmt.Println("Push endpoint:", endpoint["EndpointAddress"])
	fmt.Println("Persistent:", endpoint["Persistent"])

	paginator := sns.NewListTopicsPaginator(client, &sns.ListTopicsInput{})
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			panic(err)
		}
		for _, listed := range page.Topics {
			fmt.Println("Topic:", aws.ToString(listed.TopicArn))
		}
	}
}
