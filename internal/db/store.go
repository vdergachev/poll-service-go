package db

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gofiber/fiber/v2/log"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bundebug"
	"os"
	"strconv"
)

// Store is fine for the test task, but I prefer to split interface into the
// repositories (PollRepository, VoteRepository and OptionRepository ...)
type Store interface {
	Init(ctx context.Context) error

	SavePoll(ctx context.Context, poll *Poll) error
	PollStats(ctx context.Context, pollID int64) ([]PollResult, error)
	IsPollExists(ctx context.Context, pollID int64) (bool, error)

	SaveVote(ctx context.Context, vote *PollVote) error
	IsVoteExists(ctx context.Context, poolID, optionID, userID int64) (bool, error)

	IsOptionExists(ctx context.Context, optionID int64) (bool, error)
}

type store struct {
	db *bun.DB
}

func NewStore() (Store, error) {

	dbHost := os.Getenv("DB_HOST")
	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	dbUsername := os.Getenv("DB_USERNAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_DATABASE")

	log.Infof("Connecting to DB %s:%d %s", dbHost, dbPort, dbName)

	var datasource = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", dbUsername, dbPassword, dbHost, dbPort, dbName)

	pgDb, opErr := sql.Open("pgx", datasource)
	if opErr != nil {
		log.Error("Failed to connect to the database:", opErr)
		return nil, opErr
	}

	db := bun.NewDB(pgDb, pgdialect.New())

	db.AddQueryHook(bundebug.NewQueryHook(
		bundebug.WithVerbose(true),
		bundebug.FromEnv("BUNDEBUG"),
	))

	return &store{db: db}, nil
}

func (s store) Init(ctx context.Context) error {
	return s.db.ResetModel(ctx,
		(*Poll)(nil),
		(*PollOption)(nil),
		(*PollVote)(nil),
	)
	//return nil
}

func (s store) SavePoll(ctx context.Context, poll *Poll) error {
	return s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		// save poll
		_, pollErr := tx.NewInsert().Model(poll).Exec(ctx)
		if pollErr != nil {
			return pollErr
		}
		// save options
		for _, op := range poll.Options {
			op.PollID = poll.ID
		}
		_, opErr := tx.NewInsert().Model(&poll.Options).Exec(ctx)
		return opErr
	})
}

func (s store) SaveVote(ctx context.Context, vote *PollVote) error {
	_, err := s.db.NewInsert().Model(vote).Exec(ctx)
	return err
}

func (s store) PollStats(ctx context.Context, pollID int64) ([]PollResult, error) {
	var results []PollResult

	// Again, not enough time to rewrite raw query in bun stile
	query := `
	SELECT po.name AS option_name, COUNT(pv.id) AS votes
	FROM poll_options po
	LEFT JOIN poll_votes pv ON po.id = pv.option_id
	WHERE po.poll_id = ?
	GROUP BY po.id, po.name
	ORDER BY votes DESC;
	`
	err := s.db.NewRaw(query, pollID).Scan(ctx, &results)

	return results, err
}

func (s store) IsPollExists(ctx context.Context, pollID int64) (bool, error) {
	return s.db.NewSelect().
		Model((*Poll)(nil)).
		Where("id = ?", pollID).
		Exists(ctx)
}

func (s store) IsVoteExists(ctx context.Context, poolID, optionID, userID int64) (bool, error) {
	return s.db.NewSelect().
		Model((*PollVote)(nil)).
		Where("id = ?", optionID).
		Where("poll_id = ?", poolID).
		Where("user_id = ?", userID).
		Exists(ctx)
}

func (s store) IsOptionExists(ctx context.Context, optionID int64) (bool, error) {
	return s.db.NewSelect().
		Model((*PollOption)(nil)).
		Where("id = ?", optionID).
		Exists(ctx)
}
