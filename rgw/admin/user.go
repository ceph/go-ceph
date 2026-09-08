package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// User is GO representation of the json output of a user creation
type User struct {
	ID                  string         `json:"user_id" url:"uid"`
	DisplayName         string         `json:"display_name" url:"display-name"`
	Email               string         `json:"email" url:"email"`
	Suspended           *int           `json:"suspended" url:"suspended"`
	MaxBuckets          *int           `json:"max_buckets" url:"max-buckets"`
	Subusers            []SubuserSpec  `json:"subusers" url:"-"`
	Keys                []UserKeySpec  `json:"keys"`
	SwiftKeys           []SwiftKeySpec `json:"swift_keys" url:"-"`
	Caps                []UserCapSpec  `json:"caps"`
	OpMask              string         `json:"op_mask" url:"op-mask"`
	DefaultPlacement    string         `json:"default_placement" url:"default-placement"`
	DefaultStorageClass string         `json:"default_storage_class" url:"default-storage-class"`
	PlacementTags       []string       `json:"placement_tags" url:"placement-tags"`
	BucketQuota         QuotaSpec      `json:"bucket_quota"`
	UserQuota           QuotaSpec      `json:"user_quota"`
	TempURLKeys         []interface{}  `json:"temp_url_keys"`
	Type                string         `json:"type"`
	MfaIds              []interface{}  `json:"mfa_ids"` //revive:disable-line:var-naming old-yet-exported public api
	KeyType             string         `url:"key-type"`
	// Tenant is sent as its own "tenant" parameter by CreateUser only.
	// GetUser, ModifyUser and RemoveUser fold a populated Tenant into the
	// uid as "tenant$uid"; combining it with an ID that already names a
	// different tenant is an error. Responses never populate this field:
	// the combined form is returned in ID.
	Tenant       string   `url:"tenant"`
	GenerateKey  *bool    `url:"generate-key"`
	PurgeData    *int     `url:"purge-data"`
	GenerateStat *bool    `url:"stats"`
	Stat         UserStat `json:"stats"`
	UserCaps     string   `url:"user-caps"`
	AccountID    string   `json:"account_id" url:"account-id"`
	AccountRoot  *bool    `url:"account-root"`
}

// SubuserSpec represents a subusers of a ceph-rgw user
type SubuserSpec struct {
	Name   string        `json:"id" url:"subuser"`
	Access SubuserAccess `json:"permissions" url:"access"`

	// these are always nil in answers, they are only relevant in requests
	GenerateKey       *bool   `json:"-" url:"generate-key"`
	GenerateAccessKey *bool   `json:"-" url:"gen-access-key"`
	AccessKey         *string `json:"-" url:"access-key"`
	SecretKey         *string `json:"-" url:"secret-key"`
	Secret            *string `json:"-" url:"secret"`
	PurgeKeys         *bool   `json:"-" url:"purge-keys"`
	KeyType           *string `json:"-" url:"key-type"`
}

// SubuserAccess represents an access level for a subuser
type SubuserAccess string

// The possible values of SubuserAccess
//
// There are two sets of constants as the API parameters and the
// values returned by the API do not match.  The SubuserAccess* values
// must be used when setting access level, the SubuserAccessReply*
// values are the ones that may be returned. This is a design problem
// of the upstream API. We do not feel confident to do the mapping in
// the library.
const (
	SubuserAccessNone      SubuserAccess = ""
	SubuserAccessRead      SubuserAccess = "read"
	SubuserAccessWrite     SubuserAccess = "write"
	SubuserAccessReadWrite SubuserAccess = "readwrite"
	SubuserAccessFull      SubuserAccess = "full"

	SubuserAccessReplyNone      SubuserAccess = "<none>"
	SubuserAccessReplyRead      SubuserAccess = "read"
	SubuserAccessReplyWrite     SubuserAccess = "write"
	SubuserAccessReplyReadWrite SubuserAccess = "read-write"
	SubuserAccessReplyFull      SubuserAccess = "full-control"
)

// SwiftKeySpec represents the secret key associated to a subuser
type SwiftKeySpec struct {
	User      string `json:"user"`
	SecretKey string `json:"secret_key"`
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
	// Request fields
	UID         string `url:"uid"`     // The user ID to receive the new key
	SubUser     string `url:"subuser"` // The subuser ID to receive the new key
	KeyType     string `url:"key-type"`
	GenerateKey *bool  `url:"generate-key"` // Generate a new key pair and add to the existing keyring
}

// UserStat contains information about storage consumption by the ceph user
type UserStat struct {
	Size        *uint64 `json:"size"`
	SizeRounded *uint64 `json:"size_rounded"`
	NumObjects  *uint64 `json:"num_objects"`
}

// withTenantUID returns a copy of u with a populated Tenant folded into
// ID as "tenant$uid", the combined form RGW resolves tenancy through on
// every operation. An ID that already names a different tenant is an
// error, as is a Tenant without any ID (lookup by access key).
func (u User) withTenantUID() (User, error) {
	if u.Tenant == "" {
		return u, nil
	}
	if u.ID == "" {
		return User{}, errTenantWithoutUserID
	}
	if i := strings.IndexByte(u.ID, '$'); i >= 0 {
		if u.ID[:i] != u.Tenant {
			return User{}, fmt.Errorf("%w: tenant %q, user ID %q", errTenantMismatch, u.Tenant, u.ID)
		}
		return u, nil
	}
	u.ID = u.Tenant + "$" + u.ID
	return u, nil
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
	user, err := user.withTenantUID()
	if err != nil {
		return User{}, err
	}

	//  valid parameters not supported by go-ceph: sync
	body, err := api.call(ctx, http.MethodGet, "/user", valueToURLParams(user, []string{"uid", "access-key", "stats"}))
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

	// valid parameters not supported by go-ceph: system, exclusive
	body, err := api.call(ctx, http.MethodPut, "/user", valueToURLParams(user, []string{"uid", "display-name", "default-placement", "placement-tags", "default-storage-class", "email", "key-type", "access-key", "secret-key", "user-caps", "tenant", "generate-key", "max-buckets", "suspended", "op-mask", "account-id", "account-root"}))
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
	user, err := user.withTenantUID()
	if err != nil {
		return err
	}

	_, err = api.call(ctx, http.MethodDelete, "/user", valueToURLParams(user, []string{"uid", "purge-data"}))
	if err != nil {
		return err
	}

	return nil
}

// ModifyUser - http://docs.ceph.com/en/latest/radosgw/adminops/#modify-user
func (api *API) ModifyUser(ctx context.Context, user User) (User, error) {
	if user.ID == "" {
		return User{}, errMissingUserID
	}
	user, err := user.withTenantUID()
	if err != nil {
		return User{}, err
	}

	// valid parameters not supported by go-ceph: system
	body, err := api.call(ctx, http.MethodPost, "/user", valueToURLParams(user, []string{"uid", "display-name", "default-placement", "placement-tags", "default-storage-class", "email", "generate-key", "access-key", "secret-key", "key-type", "max-buckets", "suspended", "op-mask", "account-id", "account-root"}))
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
