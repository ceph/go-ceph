package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// User is GO representation of the json output of a user creation
type User struct {
	ID                  string        `json:"user_id" url:"uid"`
	DisplayName         string        `json:"display_name" url:"display-name"`
	Email               string        `json:"email" url:"email"`
	Suspended           *int          `json:"suspended" url:"suspended"`
	MaxBuckets          *int          `json:"max_buckets" url:"max-buckets"`
	Subusers            []interface{} `json:"subusers"`
	Keys                []UserKeySpec `json:"keys"`
	SwiftKeys           []interface{} `json:"swift_keys"`
	Caps                []UserCapSpec `json:"caps"`
	OpMask              string        `json:"op_mask"`
	DefaultPlacement    string        `json:"default_placement"`
	DefaultStorageClass string        `json:"default_storage_class"`
	PlacementTags       []interface{} `json:"placement_tags"`
	BucketQuota         QuotaSpec     `json:"bucket_quota"`
	UserQuota           QuotaSpec     `json:"user_quota"`
	TempURLKeys         []interface{} `json:"temp_url_keys"`
	Type                string        `json:"type"`
	MfaIds              []interface{} `json:"mfa_ids"`
	KeyType             string        `url:"key-type"`
	Tenant              string        `url:"tenant"`
	GenerateKey         *bool         `url:"generate-key"`
	PurgeData           *int          `url:"purge-data"`
	GenerateStat        *bool         `url:"stats"`
	Stat                UserStat      `json:"stats"`
	UserCaps            string        `url:"user-caps"`
}

// UserCapSpec represents a user capability which gives access to certain ressources
type UserCapSpec struct {
	Type string `json:"type"`
	Perm string `json:"perm"`
}

// UserKeySpec is the user credential configuration
type UserKeySpec struct {
	User      string `json:"user"`
	AccessKey string `json:"access_key" url:"access-key"`
	SecretKey string `json:"secret_key" url:"secret-key"`
}

// UserStat contains information about storage consumption by the ceph user
type UserStat struct {
	Size        *uint64 `json:"size"`
	SizeRounded *uint64 `json:"size_rounded"`
	NumObjects  *uint64 `json:"num_objects"`
}

// GetUser retrieves a given object store user
func (api *API) GetUser(ctx context.Context, user User) (User, error) {
	if user.ID == "" && len(user.Keys) == 0 {
		return User{}, errMissingUserID
	}
	if len(user.Keys) > 0 {
		for _, key := range user.Keys {
			if key.AccessKey == "" {
				return User{}, errMissingUserAccessKey
			}
		}
	}

	body, err := api.call(ctx, http.MethodGet, "/user", valueToURLParams(user))
	if err != nil {
		return User{}, err
	}

	u := User{}
	err = json.Unmarshal(body, &u)
	if err != nil {
		return User{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return u, nil
}

// GetUsers lists all object store users
func (api *API) GetUsers(ctx context.Context) (*[]string, error) {
	body, err := api.call(ctx, http.MethodGet, "/metadata/user", nil)
	if err != nil {
		return nil, err
	}
	var users *[]string
	err = json.Unmarshal(body, &users)
	if err != nil {
		return nil, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return users, nil
}

// CreateUser creates a user in the object store
func (api *API) CreateUser(ctx context.Context, user User) (User, error) {
	if user.ID == "" {
		return User{}, errMissingUserID
	}
	if user.DisplayName == "" {
		return User{}, errMissingUserDisplayName
	}

	// Send request
	body, err := api.call(ctx, http.MethodPut, "/user", valueToURLParams(user))
	if err != nil {
		return User{}, err
	}

	// Unmarshal response into Go type
	u := User{}
	err = json.Unmarshal(body, &u)
	if err != nil {
		return User{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return u, nil
}

// RemoveUser remove an user from the object store
func (api *API) RemoveUser(ctx context.Context, user User) error {
	if user.ID == "" {
		return errMissingUserID
	}

	_, err := api.call(ctx, http.MethodDelete, "/user", valueToURLParams(user))
	if err != nil {
		return err
	}

	return nil
}

// ModifyUser - http://docs.ceph.com/docs/latest/radosgw/adminops/#modify-user
func (api *API) ModifyUser(ctx context.Context, user User) (User, error) {
	if user.ID == "" {
		return User{}, errMissingUserID
	}

	body, err := api.call(ctx, http.MethodPost, "/user", valueToURLParams(user))
	if err != nil {
		return User{}, err
	}

	u := User{}
	err = json.Unmarshal(body, &u)
	if err != nil {
		return User{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return u, nil
}
