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
		// luminous adds client specific bans, only consider whole node ban ending in "0/0"
		blacklistIp := strings.Split(node.Addr, ":")
		if blacklistIp[0] == ip && blacklistIp[1] == "0/0" {
			return true
		}
	}
	return false
}
