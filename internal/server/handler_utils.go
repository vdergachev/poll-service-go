package server

import (
	"fmt"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"log"
	"strconv"
)

func badRequest(c *fiber.Ctx, msg string, err error) error {
	return sendError(c, fiber.StatusBadRequest, msg, err)
}

// We should support the list of errors in the response body, not the single one
func sendError(c *fiber.Ctx, code int, msg string, err error) error {
	return c.Status(code).
		JSON(fiber.Map{
			"status":  "error",
			"message": msg,
			"error":   err,
		})
}

func paramInt64(c *fiber.Ctx, param string) (int64, error) {
	value, err := strconv.ParseInt(c.Params(param), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %s", param, err.Error())
	}
	return value, nil
}

func wsParamInt64(c *websocket.Conn, param string) (int64, error) {
	value, err := strconv.ParseInt(c.Params(param), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter: %s", param, err.Error())
	}
	return value, nil
}

func wsSendError(c *websocket.Conn, msg string, err error) {

	wrErr := c.WriteJSON(fiber.Map{
		"status":  "error",
		"message": msg,
		"error":   err,
	})

	if wrErr != nil {
		log.Println(wrErr)
	}

	if clErr := c.Close(); clErr != nil {
		log.Println(clErr)
	}
}
