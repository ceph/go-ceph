package api

type CephNodes struct {
	Children []int  `json:"children"`
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	TypeID   int    `json:"type_id"`
	CrushWeight     float64 `json:"crush_weight"`
	Depth           int     `json:"depth"`
	Exists          int     `json:"exists"`
	PrimaryAffinity float64     `json:"primary_affinity"`
	Reweight        float64     `json:"reweight"`
	Status          string  `json:"status"`
}

type OsdTree struct {
	Output struct {
		Nodes []CephNodes `json:"nodes"`
		Stray []interface{} `json:"stray"`
	} `json:"output"`
	Status string `json:"status"`
}
