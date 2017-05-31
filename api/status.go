package api

type Status struct {
	Output struct {
		ElectionEpoch int    `json:"election_epoch"`
		Fsid          string `json:"fsid"`
		Fsmap         struct {
			ByRank []struct {
				FilesystemID int    `json:"filesystem_id"`
				Name         string `json:"name"`
				Rank         int    `json:"rank"`
				Status       string `json:"status"`
			} `json:"by_rank"`
			Epoch      int `json:"epoch"`
			ID         int `json:"id"`
			In         int `json:"in"`
			Max        int `json:"max"`
			Up         int `json:"up"`
			Up_standby int `json:"up:standby"`
		} `json:"fsmap"`
		Health struct {
			Detail []interface{} `json:"detail"`
			Health struct {
				HealthServices []struct {
					Mons []struct {
						AvailPercent int    `json:"avail_percent"`
						Health       string `json:"health"`
						KbAvail      int    `json:"kb_avail"`
						KbTotal      int    `json:"kb_total"`
						KbUsed       int    `json:"kb_used"`
						LastUpdated  string `json:"last_updated"`
						Name         string `json:"name"`
						StoreStats   struct {
							BytesLog    int    `json:"bytes_log"`
							BytesMisc   int    `json:"bytes_misc"`
							BytesSst    int    `json:"bytes_sst"`
							BytesTotal  int    `json:"bytes_total"`
							LastUpdated string `json:"last_updated"`
						} `json:"store_stats"`
					} `json:"mons"`
				} `json:"health_services"`
			} `json:"health"`
			OverallStatus string    `json:"overall_status"`
			Summary       []Summary `json:"summary"`
			Timechecks    struct {
				Epoch int `json:"epoch"`
				Mons  []struct {
					Health  string  `json:"health"`
					Latency float64 `json:"latency"`
					Name    string  `json:"name"`
					Skew    float64 `json:"skew"`
				} `json:"mons"`
				Round       int    `json:"round"`
				RoundStatus string `json:"round_status"`
			} `json:"timechecks"`
		} `json:"health"`
		Monmap struct {
			Created  string    `json:"created"`
			Epoch    int       `json:"epoch"`
			Fsid     string    `json:"fsid"`
			Modified string    `json:"modified"`
			Mons     []Monitor `json:"mons"`
		} `json:"monmap"`
		Osdmap struct {
			Osdmap struct {
				Epoch          int  `json:"epoch"`
				Full           bool `json:"full"`
				Nearfull       bool `json:"nearfull"`
				NumInOsds      int  `json:"num_in_osds"`
				NumOsds        int  `json:"num_osds"`
				NumRemappedPgs int  `json:"num_remapped_pgs"`
				NumUpOsds      int  `json:"num_up_osds"`
			} `json:"osdmap"`
		} `json:"osdmap"`
		Pgmap struct {
			BytesAvail int `json:"bytes_avail"`
			BytesTotal int `json:"bytes_total"`
			BytesUsed  int `json:"bytes_used"`
			DataBytes  int `json:"data_bytes"`
			NumPgs     int `json:"num_pgs"`
			PgsByState []struct {
				Count     int    `json:"count"`
				StateName string `json:"state_name"`
			} `json:"pgs_by_state"`
			ReadBytesSec  int `json:"read_bytes_sec"`
			ReadOpPerSec  int `json:"read_op_per_sec"`
			Version       int `json:"version"`
			WriteBytesSec int `json:"write_bytes_sec"`
			WriteOpPerSec int `json:"write_op_per_sec"`
		} `json:"pgmap"`
		Quorum      []int    `json:"quorum"`
		QuorumNames []string `json:"quorum_names"`
	} `json:"output"`
	Status string `json:"status"`
}

type Monitor struct {
	Addr string `json:"addr"`
	Name string `json:"name"`
	Rank int    `json:"rank"`
}
type Summary struct {
	Summary  string `json:"summary"`
	Severity string `json:"severity"`
}
