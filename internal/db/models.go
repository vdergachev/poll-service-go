package db

import (
	"github.com/uptrace/bun"
	"time"
)

type Poll struct {
	bun.BaseModel `bun:"table:polls,alias:p"`

	ID          int64     `bun:",pk,autoincrement"`
	Name        string    `bun:"name,notnull,type:text"`
	Description string    `bun:"description,type:text"`
	CreatedAt   time.Time `bun:"created_at,notnull,default:current_timestamp"`

	Options []*PollOption `bun:"rel:has-many,join:id=poll_id"`
}

type PollOption struct {
	bun.BaseModel `bun:"table:poll_options,alias:po"`

	ID          int64  `bun:",pk,autoincrement"`
	PollID      int64  `bun:"poll_id,notnull,type:bigint"`
	Name        string `bun:"name,notnull,type:text"`
	Description string `bun:"description,type:text"`

	Poll  *Poll       `bun:"rel:belongs-to,join:poll_id=id"`
	Votes []*PollVote `bun:"rel:has-many,join:id=option_id"`
}

type PollVote struct {
	bun.BaseModel `bun:"table:poll_votes,alias:pv"`

	ID       int64 `bun:",pk,autoincrement"`
	UserID   int64 `bun:"user_id,notnull,type:bigint"`
	PollID   int64 `bun:"poll_id,notnull,type:bigint"`
	OptionID int64 `bun:"option_id,notnull,type:bigint"`

	PollOption *PollOption `bun:"rel:belongs-to,join:option_id=id"`
	Poll       *Poll       `bun:"rel:belongs-to,join:poll_id=id"`
}

type PollResult struct {
	Option string `bun:"option_name"`
	Votes  int64  `bun:"votes"`
}
