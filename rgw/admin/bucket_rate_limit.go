//go:build !(pacific || quincy) && ceph_preview

package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SetIndividualBucketRateLimit sets rate-limits to a specific bucket.
// https://docs.ceph.com/en/latest/radosgw/adminops/#set-rate-limit-for-an-individual-bucket
func (api *API) SetIndividualBucketRateLimit(ctx context.Context, limits RateLimitSpec) error {
	if limits.Bucket == "" {
		return errMissingUserBucket
	}

	_, err := api.call(ctx, http.MethodPost, "/ratelimit?ratelimit-scope=bucket",
		valueToURLParams(limits, []string{
			"bucket",
			"enabled",
			"max-read-ops",
			"max-write-ops",
			"max-read-bytes",
			"max-write-bytes",
			"max-list-ops",
			"max-delete-ops",
		}))
	if err != nil {
		return err
	}

	return nil
}

// GetIndividualBucketRateLimit returns rate-limits of a specific bucket.
// https://docs.ceph.com/en/latest/radosgw/adminops/#get-bucket-rate-limit
func (api *API) GetIndividualBucketRateLimit(ctx context.Context, bucket string) (RateLimitSpec, error) {
	if bucket == "" {
		return RateLimitSpec{}, errMissingUserBucket
	}

	body, err := api.call(ctx, http.MethodGet, "/ratelimit?ratelimit-scope=bucket",
		valueToURLParams(RateLimitSpec{Bucket: bucket}, []string{"bucket"}))
	if err != nil {
		return RateLimitSpec{}, err
	}

	ref := struct {
		RateLimitSpec `json:"bucket_ratelimit"`
	}{}

	err = json.Unmarshal(body, &ref)
	if err != nil {
		return RateLimitSpec{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	ref.RateLimitSpec.Bucket = bucket
	return ref.RateLimitSpec, nil
}
