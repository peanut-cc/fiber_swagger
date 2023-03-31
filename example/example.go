package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/peanut-cc/fiber_swagger"
	"net/http"
)

func Success(c *fiber.Ctx, v interface{}) error {
	return ReturnJson(c, http.StatusOK, v)
}

func ReturnJson(c *fiber.Ctx, status int, v interface{}) error {
	c.Status(status)
	return c.JSON(v)
}

type FailedResponse struct {
	Err string `json:"err" description:"错误信息"`
	Msg string `json:"msg" description:"错误描述"`
}

func Unauthorized(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusUnauthorized)
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusNoContent)
}

func Forbidden(c *fiber.Ctx) error {
	return c.SendStatus(http.StatusForbidden)
}

type User struct {
	Name string `json:"name" description:"名字"`
	Age  int    `json:"age" description:"年龄"`
}

type Users []User

type QueryUsersResult struct {
	Data Users `json:"data"`
}

func QueryUsers(c *fiber.Ctx) error {
	users := Users{
		User{
			Name: "Peanut",
			Age:  12,
		},
		User{
			Name: "fan-tastic",
			Age:  18,
		},
	}
	return Success(c, users)
}

func initSwagger() *fiber_swagger.Swagger {
	swag := fiber_swagger.NewSwagger(
		"fiber_swagger example",
		"fiber_swagger generate openapi document",
		"0.0.1",
		fiber_swagger.DocPath("./docs/openapi.yaml"),
	)
	return swag
}

func main() {
	app := fiber.New()
	swag := initSwagger()

	api := app.Group("api")

	user := api.Group("user")

	user.Post("/query_users", QueryUsers).Name("查询用户")
	swag.Bind("查询用户", nil, map[int]interface{}{
		http.StatusOK:                  &QueryUsersResult{},
		http.StatusInternalServerError: &FailedResponse{},
		http.StatusNoContent:           nil,
		http.StatusUnauthorized:        nil,
		http.StatusForbidden:           nil,
		http.StatusBadRequest:          &FailedResponse{},
	})

	swag.Generate(app)
	//app.Listen(":3000")
}
