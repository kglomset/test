package domain

type TrackConditions struct {
	ID            int    `json:"id"`
	TrackHardness string `json:"track_hardness"` // 'H1', 'H2', 'H3', 'H4', 'H5', 'H6'
	TrackType     string `json:"track_type"`     // 'T1', 'T2', 'D1', 'D2'
}
