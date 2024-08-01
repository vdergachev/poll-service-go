package server

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"voting-service/internal/db"
	rds "voting-service/internal/redis"
)

type voteHandler struct {
	store db.Store
	rd    *redis.Client
}

func NewVoteHandler(store db.Store, rd *redis.Client) RestHandler {
	return voteHandler{
		store: store,
		rd:    rd,
	}
}

func (h voteHandler) Handle(c *fiber.Ctx) error {

	pollId, pollErr := paramInt64(c, "pollId")
	if pollErr != nil {
		return badRequest(c, "Invalid pollId request param", pollErr)
	}

	optionId, optionErr := paramInt64(c, "optionId")
	if optionErr != nil {
		return badRequest(c, "Invalid optionId request param", optionErr)
	}

	userId, userErr := paramInt64(c, "userId")
	if userErr != nil {
		return badRequest(c, "Invalid userId request param", userErr)
	}

	if validateErr := h.validate(c.Context(), pollId, optionId, userId); validateErr != nil {
		return badRequest(c, "Validation failure", validateErr)
	}

	vote := h.createVote(pollId, optionId, userId)

	saveErr := h.store.SaveVote(c.Context(), vote)
	if saveErr != nil {
		return badRequest(c, "Database failure", saveErr)
	}

	// And distribute them across service instances
	// Should be wrapped into transaction with SaveVote
	if err := h.publishVoteEvent(c, vote); err != nil {
		return badRequest(c, "Redis failure", err)
	}

	return c.Status(http.StatusNoContent).JSON("")
}

func (h voteHandler) publishVoteEvent(c *fiber.Ctx, vote *db.PollVote) error {

	event := rds.NewVote(
		vote.UserID,
		vote.PollID,
		vote.OptionID,
	)

	cmd := h.rd.Publish(c.Context(), rds.VoteChannel, event)
	return cmd.Err()
}

func (h voteHandler) validate(ctx context.Context, poolID, optionID, userID int64) error {

	// Check poll existence
	if pollExists, pollErr := h.store.IsPollExists(ctx, poolID); pollErr != nil {
		return pollErr
	} else if !pollExists {
		return errors.New("poll does not exist")
	}

	// Check option existence
	if optExists, optErr := h.store.IsOptionExists(ctx, optionID); optErr != nil {
		return optErr
	} else if !optExists {
		return errors.New("option for poll does not exist")
	}

	// Check vote is done
	if voteExists, voteErr := h.store.IsVoteExists(ctx, poolID, optionID, userID); voteErr != nil {
		return voteErr
	} else if voteExists {
		return errors.New("vote for this poll already done")
	}

	return nil
}

func (h voteHandler) createVote(poolID, optionID, userID int64) *db.PollVote {
	vote := &db.PollVote{
		UserID:   userID,
		PollID:   poolID,
		OptionID: optionID,
	}
	return vote
}
