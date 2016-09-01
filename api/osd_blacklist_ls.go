package api

type OsdBlacklistLs struct {
	Nodes []struct {
		Addr  string `json:"addr"`
		Until string `json:"until"`
	} `json:"output"`
	Status string `json:"status"`
}
