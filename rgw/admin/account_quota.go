//go:build !(pacific || quincy || reef) && ceph_preview

package admin

import (
	"context"
	"net/http"
)

// Account quota types accepted by SetAccountQuota.
const (
	// AccountQuotaTypeAccount applies the quota to the account as a whole.
	AccountQuotaTypeAccount = "account"
	// AccountQuotaTypeBucket applies the quota as the default per-bucket quota
	// for buckets owned by the account.
	AccountQuotaTypeBucket = "bucket"
)

// AccountQuotaSpec describes a quota to set on an RGW account. QuotaType selects
// whether the limits apply to the account as a whole (AccountQuotaTypeAccount)
// or as the default quota for each bucket the account owns
// (AccountQuotaTypeBucket). Unset (nil) fields are left unchanged.
type AccountQuotaSpec struct {
	ID         string `url:"id"`
	QuotaType  string `url:"quota-type"`
	Enabled    *bool  `url:"enabled"`
	MaxSize    *int64 `url:"max-size"`
	MaxObjects *int64 `url:"max-objects"`
}

// SetAccountQuota sets the account-level or default-bucket quota for an RGW
// account. The resulting quota can be read back with GetAccount, which populates
// Account.Quota and Account.BucketQuota.
// The underlying admin op requires Ceph v20.2.2 or later, or main; it is not
// available in Squid, nor in Tentacle before v20.2.2.
// https://docs.ceph.com/en/latest/radosgw/adminops/#set-account-quota
func (api *API) SetAccountQuota(ctx context.Context, quota AccountQuotaSpec) error {
	if quota.ID == "" {
		return ErrInvalidArgument
	}
	if quota.QuotaType != AccountQuotaTypeAccount && quota.QuotaType != AccountQuotaTypeBucket {
		return ErrInvalidArgument
	}

	_, err := api.call(ctx, http.MethodPut, "/account?quota", valueToURLParams(quota, []string{"id", "quota-type", "enabled", "max-size", "max-objects"}))
	if err != nil {
		return err
	}

	return nil
}
