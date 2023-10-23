package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// SetIndividualBucketQuota sets quota to a specific bucket
// https://docs.ceph.com/en/latest/radosgw/adminops/#set-quota-for-an-individual-bucket
func (api *API) SetIndividualBucketQuota(ctx context.Context, quota QuotaSpec) error {
	if quota.UID == "" {
		return errMissingUserID
	}

	if quota.Bucket == "" {
		return errMissingUserBucket
	}

	_, err := api.call(ctx, http.MethodPut, "/bucket?quota", valueToURLParams(quota, []string{"bucket", "uid", "enabled", "max-size", "max-size-kb", "max-objects"}))
	if err != nil {
		return err
	}

	return nil
}

// GetBucketQuota will return the bucket quota for a user
func (api *API) GetBucketQuota(ctx context.Context, quota QuotaSpec) (QuotaSpec, error) {
	// Always for quota type to bucket
	quota.QuotaType = "bucket"

	if quota.UID == "" {
		return QuotaSpec{}, errMissingUserID
	}

	body, err := api.call(ctx, http.MethodGet, "/user?quota", valueToURLParams(quota, []string{"uid", "quota-type"}))
	if err != nil {
		return QuotaSpec{}, err
	}

	ref := QuotaSpec{}
	err = json.Unmarshal(body, &ref)
	if err != nil {
		return QuotaSpec{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return ref, nil
}

// SetBucketQuota sets quota to a user
// Global quotas (https://docs.ceph.com/en/latest/radosgw/admin/#reading-writing-global-quotas) are not surfaced in the Admin Ops API
// So this library cannot expose it yet
func (api *API) SetBucketQuota(ctx context.Context, quota QuotaSpec) error {
	// Always for quota type to bucket
	quota.QuotaType = "bucket"

	if quota.UID == "" {
		return errMissingUserID
	}

	_, err := api.call(ctx, http.MethodPut, "/user?quota", valueToURLParams(quota, []string{"uid", "quota-type", "enabled", "max-size", "max-size-kb", "max-objects"}))
	if err != nil {
		return err
	}

	return nil
}
