package redis

import (
	"encoding/json"
)

const (
	VoteChannel = "poll_votes"
)

type Vote struct {
	UserID   int64
	PollID   int64
	OptionID int64
}

func NewVote(userID, pollID, optionID int64) *Vote {
	return &Vote{UserID: userID, PollID: pollID, OptionID: optionID}
}

func (pv Vote) MarshalBinary() (data []byte, err error) {
	bytes, err := json.Marshal(pv)
	return bytes, err
}
