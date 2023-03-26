# Fiber Swagger

Currently in the development stage

## Introduction

Most of the current golang web framework OpenAPI document generation is implemented by way of comments, which feels too bloated for business code, fiber_swagger automatically generates OpenAPI documents by adding code to the route and tag fields in the request and response data structures


## example

```go
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

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
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
	swag.Bind("查询用户", nil, &QueryUsersResult{})

	swag.Generate(app)
	//app.Listen(":3000")
}
```
