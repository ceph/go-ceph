//go:build ceph_preview && !squid

package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Account represents an RGW account
type Account struct {
	ID            string    `json:"id" url:"id"`
	Name          string    `json:"name" url:"name"`
	Email         string    `json:"email" url:"email"`
	Tenant        string    `json:"tenant" url:"tenant"`
	MaxUsers      *int64    `json:"max_users" url:"max-users"`
	MaxRoles      *int64    `json:"max_roles" url:"max-roles"`
	MaxGroups     *int64    `json:"max_groups" url:"max-groups"`
	MaxAccessKeys *int64    `json:"max_access_keys" url:"max-access-keys"`
	MaxBuckets    *int64    `json:"max_buckets" url:"max-buckets"`
	Quota         QuotaSpec `json:"quota"`
	BucketQuota   QuotaSpec `json:"bucket_quota"`
}

// CreateAccount will create a new RGW account
func (api *API) CreateAccount(ctx context.Context, account Account) (Account, error) {

	body, err := api.call(ctx, http.MethodPost, "/account", valueToURLParams(account, []string{"name", "email", "tenant", "max-users", "max-roles", "max-groups", "max-access-keys", "max-buckets"}))
	if err != nil {
		return Account{}, err
	}

	a := Account{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return Account{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return a, nil
}

// GetAccount will return the RGW account details
func (api *API) GetAccount(ctx context.Context, accountID string) (Account, error) {
	if accountID == "" {
		return Account{}, ErrInvalidArgument
	}

	body, err := api.call(ctx, http.MethodGet, "/account", valueToURLParams(Account{ID: accountID}, []string{"id"}))
	if err != nil {
		return Account{}, err
	}

	a := Account{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return Account{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return a, nil
}

// DeleteAccount will delete the RGW account
func (api *API) DeleteAccount(ctx context.Context, accountID string) error {
	if accountID == "" {
		return ErrInvalidArgument
	}

	_, err := api.call(ctx, http.MethodDelete, "/account", valueToURLParams(Account{ID: accountID}, []string{"id"}))
	if err != nil {
		return err
	}

	return nil
}

// ModifyAccount will modify the RGW account
func (api *API) ModifyAccount(ctx context.Context, account Account) (Account, error) {
	if account.ID == "" {
		return Account{}, ErrInvalidArgument
	}

	body, err := api.call(ctx, http.MethodPut, "/account", valueToURLParams(account, []string{"id", "name", "email", "tenant", "max-users", "max-roles", "max-groups", "max-access-keys", "max-buckets"}))
	if err != nil {
		return Account{}, err
	}

	a := Account{}
	err = json.Unmarshal(body, &a)
	if err != nil {
		return Account{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return a, nil
}
