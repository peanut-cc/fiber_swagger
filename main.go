package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/peanut-cc/fiberx/swagger"
)

//type QueryNamesArgs struct {
//	Id string `json:"id" description:"id"`
//}
//
//type QueryNameResult struct {
//	Names []string `json:"names" description:"名字列表"`
//}
//
//func QueryName(c *fiber.Ctx) error {
//	names := &QueryNameResult{Names: []string{"Golang", "Rust"}}
//	return Success(c, names)
//}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Users []User

type QueryUsersResult struct {
	Data Users `embed:"" json:"data"`
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

func main() {
	app := fiber.New()
	r := swagger.NewSwagger()

	api := app.Group("api")

	user := api.Group("user")

	//user.Get("/", QueryName).Name("查询用户")
	//r.Bind("查询用户", &QueryNamesArgs{}, &QueryNameResult{})
	user.Get("/", QueryUsers).Name("查询用户")
	r.Bind("查询用户", nil, &QueryUsersResult{})

	r.Generate(app)
	app.Listen(":3000")
}
