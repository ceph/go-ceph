package api

type MdsStat struct {
	Output struct {
		Fsmap struct {
			Compat struct {
				Compat   struct{} `json:"compat"`
				Incompat struct {
					Feature1 string `json:"feature_1"`
					Feature2 string `json:"feature_2"`
					Feature3 string `json:"feature_3"`
					Feature4 string `json:"feature_4"`
					Feature5 string `json:"feature_5"`
					Feature6 string `json:"feature_6"`
					Feature8 string `json:"feature_8"`
				} `json:"incompat"`
				RoCompat struct{} `json:"ro_compat"`
			} `json:"compat"`
			Epoch        int `json:"epoch"`
			FeatureFlags struct {
				EnableMultiple      bool `json:"enable_multiple"`
				EverEnabledMultiple bool `json:"ever_enabled_multiple"`
			} `json:"feature_flags"`
			Filesystems []struct {
				ID     int `json:"id"`
				Mdsmap struct {
					Compat struct {
						Compat   struct{} `json:"compat"`
						Incompat struct {
							Feature1 string `json:"feature_1"`
							Feature2 string `json:"feature_2"`
							Feature3 string `json:"feature_3"`
							Feature4 string `json:"feature_4"`
							Feature5 string `json:"feature_5"`
							Feature6 string `json:"feature_6"`
							Feature8 string `json:"feature_8"`
						} `json:"incompat"`
						RoCompat struct{} `json:"ro_compat"`
					} `json:"compat"`
					Created                   string               `json:"created"`
					Damaged                   []interface{}        `json:"damaged"`
					DataPools                 []int                `json:"data_pools"`
					Enabled                   bool                 `json:"enabled"`
					Epoch                     int                  `json:"epoch"`
					EverAllowedFeatures       int                  `json:"ever_allowed_features"`
					ExplicitlyAllowedFeatures int                  `json:"explicitly_allowed_features"`
					Failed                    []interface{}        `json:"failed"`
					Flags                     int                  `json:"flags"`
					FsName                    string               `json:"fs_name"`
					In                        []int                `json:"in"`
					Info                      map[string]ActiveMds `json:"info"`
					LastFailure               int                  `json:"last_failure"`
					LastFailureOsdEpoch       int                  `json:"last_failure_osd_epoch"`
					MaxFileSize               int                  `json:"max_file_size"`
					MaxMds                    int                  `json:"max_mds"`
					MetadataPool              int                  `json:"metadata_pool"`
					Modified                  string               `json:"modified"`
					Root                      int                  `json:"root"`
					SessionAutoclose          int                  `json:"session_autoclose"`
					SessionTimeout            int                  `json:"session_timeout"`
					Stopped                   []interface{}        `json:"stopped"`
					Tableserver               int                  `json:"tableserver"`
					Up                        struct {
						Mds0 int `json:"mds_0"`
					} `json:"up"`
				} `json:"mdsmap"`
			} `json:"filesystems"`
			Standbys []ActiveMds `json:"standbys"`
		} `json:"fsmap"`
		MdsmapFirstCommitted int `json:"mdsmap_first_committed"`
		MdsmapLastCommitted  int `json:"mdsmap_last_committed"`
	} `json:"output"`
	Status string `json:"status"`
}

type ActiveMds struct {
	Addr            string        `json:"addr"`
	ExportTargets   []interface{} `json:"export_targets"`
	Features        int           `json:"features"`
	Gid             int           `json:"gid"`
	Incarnation     int           `json:"incarnation"`
	Name            string        `json:"name"`
	Rank            int           `json:"rank"`
	StandbyForFscid int           `json:"standby_for_fscid"`
	StandbyForName  string        `json:"standby_for_name"`
	StandbyForRank  int           `json:"standby_for_rank"`
	StandbyReplay   bool          `json:"standby_replay"`
	State           string        `json:"state"`
	StateSeq        int           `json:"state_seq"`
}
