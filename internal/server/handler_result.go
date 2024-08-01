package server

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/contrib/websocket"
	"log"
	"sync"
	"voting-service/internal/api"
	"voting-service/internal/db"
	rds "voting-service/internal/redis"
)

type resultHandler struct {
	store    db.Store
	rd       *redis.Client
	mutex    *sync.RWMutex
	channels []chan rds.Vote
	eventCh  chan rds.Vote
	chanCh   chan chan rds.Vote
}

func NewResultHandler(store db.Store, rd *redis.Client) WSHandler {

	// TODO: Calc stats for all polls on the start

	h := resultHandler{
		store: store,
		rd:    rd,

		mutex:    &sync.RWMutex{},                // channels lock
		channels: make([]chan rds.Vote, 0, 1024), // default capacity is good question ))

		eventCh: make(chan rds.Vote), // to spread events across all connection

		chanCh: make(chan chan rds.Vote), // make it buffered allow handle a lot incoming connection
	}

	// listen redis events
	go h.subscribe()

	// receive incoming connection channels
	go h.connector()

	// spread events over connection
	go h.multiplexer()

	return &h
}

func (h *resultHandler) Handle(c *websocket.Conn) {

	pollID, pollErr := wsParamInt64(c, "pollId")
	if pollErr != nil {
		wsSendError(c, "pollID is invalid", pollErr)
		return
	}

	// Use cache here to avoid DB round trip
	if isExists, exErr := h.store.IsPollExists(context.Background(), pollID); exErr != nil {
		wsSendError(c, "Database failure", pollErr) // bad practice demonstration - to expose our errors outside ^_^
		return
	} else if !isExists {
		wsSendError(c, "pollID is not found", pollErr)
		return
	}

	// Create chan per connection
	conChan := make(chan rds.Vote, 16) // Nice to have any buffer here - each event handling and send takes time
	defer close(conChan)               // I do not handle connection close, have to remove the channel from h.channels

	select {
	case h.chanCh <- conChan: // send this cha to the connector and multiplexer routines
		log.Println("new connection chan sent to chan-chan")
		break
	default:
		wsSendError(c, "Server is busy, try later", pollErr)
		return
	}

	log.Println("connection waiting for vote")

	for event := range conChan {

		if event.PollID != pollID {
			continue
		}

		// TODO: Update poll stats here, do not count like I do

		stats, statsErr := h.store.PollStats(context.Background(), event.PollID)
		if statsErr != nil {
			wsSendError(c, "Database failure", statsErr) // bad practice
		}

		response := h.createResponse(pollID, stats)

		// Send poll stats to client
		wrErr := c.Conn.WriteJSON(response)
		if wrErr != nil {
			log.Println("Write error:", wrErr)
			break
		}
	}

}

func (h *resultHandler) subscribe() {

	log.Println("subscribe started to work")

	s := h.rd.Subscribe(context.Background(), rds.VoteChannel)
	ch := s.Channel()

	for msg := range ch {

		var voteEvent rds.Vote
		parseErr := json.Unmarshal([]byte(msg.Payload), &voteEvent)
		if parseErr != nil {
			log.Println("Unmarshal error:", parseErr)
			continue
		}

		select {
		case h.eventCh <- voteEvent:
			log.Println("vote sent to event channel")
		default:
			log.Println("vote channel is full or no consumers connected yet")
		}
	}

	log.Println("subscriber finished to work")
}

func (h *resultHandler) multiplexer() {
	log.Println("multiplexer started to work")
	for event := range h.eventCh {
		log.Println("multiplexer :: new event received: ", event)
		h.mutex.RLock()
		for _, ch := range h.channels {
			select {
			case ch <- event:
				log.Println("multiplexer :: event sent to channel")
				break
			default:
				log.Println("multiplexer :: channel is busy")
			}
		}
		h.mutex.RUnlock()
		log.Println("multiplexer :: event processed:", event)
	}
	log.Println("multiplexer finished to work")
}

func (h *resultHandler) connector() {
	log.Println("connector started to work")
	for ch := range h.chanCh {
		h.mutex.Lock()
		h.channels = append(h.channels, ch)
		h.mutex.Unlock()
		log.Println("connector :: new connection added: ")
	}
	log.Println("connector finished to work")
}

func (h *resultHandler) createResponse(id int64, stats []db.PollResult) api.PollStats {
	var total int64
	result := make([]api.PollStat, 0, len(stats))
	for _, stat := range stats {
		total += stat.Votes
		result = append(result, api.PollStat{Option: stat.Option, Votes: stat.Votes})
	}
	return api.PollStats{
		PoolID:     id,
		TotalVotes: total,
		Stats:      result,
	}
}
