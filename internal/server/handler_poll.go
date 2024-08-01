package server

import (
	"fmt"
	"net/http"
	"voting-service/internal/api"
	"voting-service/internal/db"

	"github.com/gofiber/fiber/v2"
)

type pollHandler struct {
	store db.Store
}

func NewPollHandler(store db.Store) RestHandler {
	return pollHandler{store: store}
}

func (h pollHandler) Handle(c *fiber.Ctx) error {

	create := new(api.CreatePollRequest)

	if parseErr := c.BodyParser(create); parseErr != nil {
		return badRequest(c, "Invalid poll create request", parseErr)
	}

	if pollErr := h.validatePoll(create); pollErr != nil {
		return badRequest(c, "Validation failure", pollErr)
	}

	poll := h.createPoll(create)

	saveErr := h.store.SavePoll(c.Context(), poll)
	if saveErr != nil {
		return badRequest(c, "Database failure", saveErr)
	}

	// Build response entity
	return c.Status(http.StatusCreated).JSON(h.createResponse(poll))
}

func (h pollHandler) validatePoll(create *api.CreatePollRequest) error {

	// Trivial validation for empty options
	if len(create.Options) == 0 {
		return fmt.Errorf("empty poll options")
	}

	// Add more validation cases here ...

	return nil
}

func (h pollHandler) createPoll(create *api.CreatePollRequest) *db.Poll {
	poll := &db.Poll{
		Name:        create.Name,
		Description: create.Description,
		Options:     make([]*db.PollOption, 0, len(create.Options)),
	}

	for _, op := range create.Options {
		poll.Options = append(poll.Options, &db.PollOption{
			Name:        op.Name,
			Description: op.Description,
		})
	}

	return poll
}

func (h pollHandler) createResponse(poll *db.Poll) api.CreatePollResponse {

	optionIDs := make([]api.OptionResponseDto, 0, len(poll.Options))
	for _, op := range poll.Options {
		optionIDs = append(optionIDs, api.OptionResponseDto{ID: op.ID})
	}

	return api.CreatePollResponse{
		ID:      poll.ID,
		Options: optionIDs,
	}
}
