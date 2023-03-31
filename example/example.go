package main

import (
	"fmt"
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
	Err string `json:"err" validate:"required" description:"错误信息"`
	Msg string `json:"msg" validate:"required" description:"错误描述"`
}

type User struct {
	Id       int `json:"id" validate:"required" description:"用户的唯一标识"`
	UserBase `embed:""`
}

type UserBase struct {
	Name string `json:"name" validate:"required" description:"名字"`
	Age  int    `json:"age" validate:"required" description:"年龄"`
}

type Users []User

type QueryUsersResult struct {
	Data Users `json:"data"`
}

func QueryUsers(c *fiber.Ctx) error {
	var users Users
	for i := 1; i < 10; i++ {
		user := User{}
		user.Id = i
		user.Age = i
		user.Name = fmt.Sprintf("user%d", i)
		users = append(users, user)
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
