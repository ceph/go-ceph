package admin

import (
	"context"
	"net/http"
)

// BucketLinkInput the bucket link/unlink input parameters
type BucketLinkInput struct {
	Bucket   string `url:"bucket" json:"bucket"`
	BucketID string `url:"bucket-id" json:"bucket_id"`
	UID      string `url:"uid" json:"uid"`
}

// UnlinkBucket unlink a bucket from a specified user
// Primarily useful for changing bucket ownership.
func (api *API) UnlinkBucket(ctx context.Context, link BucketLinkInput) error {
	if link.UID == "" {
		return errMissingUserID
	}
	if link.Bucket == "" {
		return errMissingBucket
	}
	_, err := api.call(ctx, http.MethodPost, "/bucket", valueToURLParams(link, []string{"uid", "bucket"}))
	return err
}

// LinkBucket will link a bucket to a specified user
// unlinking the bucket from any previous user
func (api *API) LinkBucket(ctx context.Context, link BucketLinkInput) error {
	if link.UID == "" {
		return errMissingUserID
	}
	if link.Bucket == "" {
		return errMissingBucket
	}
	//  valid parameters not supported by go-ceph: new-bucket-name
	_, err := api.call(ctx, http.MethodPut, "/bucket", valueToURLParams(link, []string{"uid", "bucket-id", "bucket"}))
	return err
}
