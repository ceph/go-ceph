package api

import (
	"strings"
)

type OsdBlacklistLs struct {
	Nodes []struct {
		Addr  string `json:"addr"`
		Until string `json:"until"`
	} `json:"output"`
	Status string `json:"status"`
}

func (obl *OsdBlacklistLs) IsInBlacklist(ip string) bool {
	for _, node := range obl.Nodes {
		if strings.Contains(node.Addr, ip) {
			return true
		}
	}
	return false
}