# Fiber Swagger

目前还处于开发阶段

## 介绍

目前大多的 `Golang` web 框架的 `OpenAPI` 文档的生成是通过写注释的方式生成的，并且注释的的内容非常多，开发者其实大多数应该不喜欢写这样的注释。`Filber_Swagger` 通过在 `Fiber` web框架路由的地方做一些额外的绑定以及在定义接口的请求和响应的结构体中增加tag来实现自动生成 `OpenAPI` 相比于 通过大量注释的方式还是比较友好的，同时`Filber_Swagger` 在尽可能减少对`Fiber` 框架的的侵入。

## 使用说明

### 快速使用

```Go
func initSwagger() *fiber_swagger.Swagger {
	swag := fiber_swagger.NewSwagger(
		"fiber_swagger example",
		"fiber_swagger generate openapi document",
		"0.0.1",
		fiber_swagger.DocPath("./docs/openapi.yaml"),
	)
	return swag
}
```

初始化 `OpenAPI` 的基本信息，包含了 Title, Description,Version，以及生成 文档的路径。

定义请求或响应的数据结构：

```Go
type User struct {
	Id       int `json:"id" validate:"required" description:"用户的唯一标识"`
	UserBase `embed:""`
}

type UserBase struct {
	Name string `json:"name" validate:"required" description:"名字"`
	Age  int    `json:"age" validate:"required" description:"年龄"`
}
```

这里是为了做嵌套嵌套数据结构的示例，所以定义了 `UserBase` 数据结构和 `User` 数据结构，当 `tag`中存在 `embed` 时会展开数据结构到当前的数据结构

做绑定关联

```Go
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
}
```

在路由的的地方通过Name指定了每个路由的名称，同时通过这个名称并绑定了请求的数据结构和响应的数据结构，关于响应的数据库结构这里是定义了一个 `map[int]interface{}` 来保存，`key` 是 HTTP 的状态码，`value` 是对应状态码的数据结构，关于不同的状态码这里原本的期望是：

- `http.StatusOK`：请求成功，并返回 `{"data": ""}` 的数据结构
- `http.StatusInternalServerError`: 表示服务器内部错误, 返回 `{"err":"",msg:""}` 的数据结构
- `http.StatusNoContent`: 表示请求成功，但不返回数据结构
- `http.StatusUnauthorized`: 表示未认证，不返回数据结构
- `http.StatusForbidden`: 表示没有权限，不返回数据结构
- `http.StatusBadRequest`: 表示错误的请求，返回 `{"err":"",msg:""}` 的数据结构

目前 关于 不同的状态码的处理，是固定的，如 `http.StatusNoContent` 在生成 `OpenAPI`的处理是固定的：

```Go
func (s *Swagger) NoContentResponse() *openapi3.ResponseRef {
	return &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Content:     nil,
			Description: &NOCONTENT,
		},
	}
}

func (s *Swagger) ForbiddenResponse() *openapi3.ResponseRef {
	return &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Content:     nil,
			Description: &FORBIDDEN,
		},
	}
}
```
