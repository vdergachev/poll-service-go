package server

import (
	"context"
	"github.com/go-redis/redis/v8"
	"log"
	"os"
	"voting-service/internal/db"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type opts struct {
	listen string
}

type RestHandler interface {
	Handle(c *fiber.Ctx) error
}

type WSHandler interface {
	Handle(c *websocket.Conn)
}

type server struct {
	app     *fiber.App
	options opts

	voteHandler   RestHandler
	poolHandler   RestHandler
	resultHandler WSHandler
}

func Run() error {

	config := opts{
		listen: os.Getenv("SERVER_LISTEN"),
	}

	serv, err := newServer(config)
	if err != nil {
		return err
	}

	serv.mountHandlers()

	return serv.run()
}

func newServer(options opts) (*server, error) {

	app := fiber.New(fiber.Config{AppName: os.Getenv("APP_NAME")})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use("/ws", webSocketUpgrade)

	// Init db
	store, storeErr := db.NewStore()
	if storeErr != nil {
		return nil, storeErr
	}

	if initErr := store.Init(context.Background()); initErr != nil {
		return nil, initErr
	}

	// Init redis
	var rd = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDRESS"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0, // use default DB
	})

	return &server{
		options: options,
		app:     app,

		voteHandler:   NewVoteHandler(store, rd),
		poolHandler:   NewPollHandler(store),
		resultHandler: NewResultHandler(store, rd),
	}, nil
}

func webSocketUpgrade(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return c.Next()
	}
	return fiber.ErrUpgradeRequired
}

func (s server) run() error {

	log.Println("Server is running now")
	return s.app.Listen(s.options.listen)
}

func (s server) mountHandlers() {

	// create poll
	s.app.Post("/polls", s.poolHandler.Handle)

	// handle vote
	s.app.Put("/polls/:pollId/options/:optionId/users/:userId", s.voteHandler.Handle)

	// spread results
	s.app.Get("/ws/polls/:pollId", websocket.New(s.resultHandler.Handle))
}
