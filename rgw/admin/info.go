package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

/* parse a nested json expression like
 * {"info":{"storage_backends":[{"name":"rados","cluster_id":"75d1938b-2949-4933-8386-fb2d1449ff03"}]}} */

// Info struct
type Backend struct {
	Name    string `json:"name"`
	ClusterID    string `json:"cluster_id"`
}

type Info struct {
	Info map[string] []Backend
}

// GetInfo fetch an array of info elements (e.g., the cluster fsid)
func (api *API) GetInfo(ctx context.Context) (Info, error) {
	body, err := api.call(ctx, http.MethodGet, "/info", nil)
	if err != nil {
		return Info{}, err
	}
	fmt.Println(string(body))
	u := Info{}
	err = json.Unmarshal(body, &u)
	if err != nil {
		return Info{}, fmt.Errorf("%s. %s. %w", unmarshalError, string(body), err)
	}

	return u, nil
}
