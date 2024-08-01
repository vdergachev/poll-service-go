package api

type CreatePollRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Options     []OptionDto `json:"options"`
}

type OptionDto struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePollResponse struct {
	ID      int64               `json:"id"`
	Options []OptionResponseDto `json:"options"`
}

type OptionResponseDto struct {
	ID int64 `json:"id"`
}

type PollStats struct {
	PoolID     int64      `json:"id"`
	TotalVotes int64      `json:"total_votes"`
	Stats      []PollStat `json:"stats"`
}

type PollStat struct {
	Option string `json:"option"`
	Votes  int64  `json:"votes"`
}
