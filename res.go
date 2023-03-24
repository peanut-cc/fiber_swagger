package main

import (
	"github.com/gofiber/fiber/v2"
	"net/http"
)

type FailedResponse struct {
	Err string `json:"err"`
	Msg string `json:"msg"`
}

func Success(c *fiber.Ctx, v interface{}) error {
	return ReturnJson(c, http.StatusOK, v)
}

func ReturnJson(c *fiber.Ctx, status int, v interface{}) error {
	c.Status(status)
	return c.JSON(v)
}
