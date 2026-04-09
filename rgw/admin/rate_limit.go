//go:build !(pacific || quincy) && ceph_preview

package admin

// RateLimitSpec describes an object store rate-limit for a user or a bucket.
type RateLimitSpec struct {
	UID           string `json:"user_id" url:"uid"`
	Bucket        string `json:"bucket" url:"bucket"`
	Enabled       *bool  `json:"enabled" url:"enabled"`
	MaxReadOps    *int64 `json:"max_read_ops" url:"max-read-ops"`
	MaxWriteOps   *int64 `json:"max_write_ops" url:"max-write-ops"`
	MaxReadBytes  *int64 `json:"max_read_bytes" url:"max-read-bytes"`
	MaxWriteBytes *int64 `json:"max_write_bytes" url:"max-write-bytes"`
	MaxListOps    *int64 `json:"max_list_ops" url:"max-list-ops"`     // Ceph 21+
	MaxDeleteOps  *int64 `json:"max_delete_ops" url:"max-delete-ops"` // Ceph 21+
}
