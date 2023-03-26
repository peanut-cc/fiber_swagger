package fiber_swagger

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/peanut-cc/fiber_swagger/router"
	"mime/multipart"
	"sync"
	"time"
)

type Swagger struct {
	routersMap      sync.Map
	Title           string
	Description     string
	Version         string
	License         *openapi3.License
	Components      []interface{}
	OpenAPI         *openapi3.T
	Schemas         map[string]*openapi3.SchemaRef
	Paths           map[string]map[string]*router.Router
	OpenAPIYamlFile string
}

func NewSwagger() *Swagger {
	openapi := &openapi3.T{
		Info:       nil,
		OpenAPI:    "3.0.0",
		Components: &openapi3.Components{},
		Tags:       openapi3.Tags{},
		Paths:      map[string]*openapi3.PathItem{},
		Security:   openapi3.SecurityRequirements{map[string][]string{"http": {}}},
	}
	return &Swagger{
		Title:           "",
		Description:     "",
		Version:         "",
		License:         nil,
		Components:      nil,
		OpenAPI:         openapi,
		Schemas:         make(map[string]*openapi3.SchemaRef),
		Paths:           make(map[string]map[string]*router.Router),
		OpenAPIYamlFile: "./docs/openapi.yaml",
	}
}

type HttpRequestResponse struct {
	URLName  string
	Request  interface{}
	Response interface{}
}

func (s *Swagger) Bind(name string, request interface{}, response interface{}) {
	httpReqRes := &HttpRequestResponse{
		URLName:  name,
		Request:  request,
		Response: response,
	}
	s.store(httpReqRes)
}

func (s *Swagger) store(args *HttpRequestResponse) {
	s.routersMap.Store(args.URLName, args)
}

func (s *Swagger) Generate(app *fiber.App) {
	for _, route := range app.GetRoutes() {
		if route.Name == "" {
			continue
		}
		reqRep := s.load(route.Name)
		req := reqRep.Request
		rep := reqRep.Response
		s.addComponents(req, rep)
		s.addPath(route, req, rep)
	}
	s.buildComponents()
	s.buildPaths()
	err := s.WriteToYaml()
	if err != nil {
		panic(err)
	}
}

func (s *Swagger) load(name string) (httpResRep *HttpRequestResponse) {
	result, _ := s.routersMap.Load(name)
	return result.(*HttpRequestResponse)
}

func (s *Swagger) getSchemaFromBaseType(field interface{}) *openapi3.Schema {
	var schema *openapi3.Schema
	var m float64
	m = float64(0)
	switch field.(type) {
	case int, int8, int16, *int, *int8, *int16:
		schema = openapi3.NewIntegerSchema()
	case uint, uint8, uint16, *uint, *uint8, *uint16:
		schema = openapi3.NewIntegerSchema()
		schema.Min = &m
	case int32, *int32:
		schema = openapi3.NewInt32Schema()
	case uint32, *uint32:
		schema = openapi3.NewInt32Schema()
		schema.Min = &m
	case int64, *int64:
		schema = openapi3.NewInt64Schema()
	case uint64, *uint64:
		schema = openapi3.NewInt64Schema()
		schema.Min = &m
	case string, *string:
		schema = openapi3.NewStringSchema()
	case time.Time, *time.Time:
		schema = openapi3.NewDateTimeSchema()
	case uuid.UUID, *uuid.UUID:
		schema = openapi3.NewUUIDSchema()
	case float32, float64, *float32, *float64:
		schema = openapi3.NewFloat64Schema()
	case bool, *bool:
		schema = openapi3.NewBoolSchema()
	case []byte:
		schema = openapi3.NewBytesSchema()
	case *multipart.FileHeader:
		schema = openapi3.NewStringSchema()
		schema.Format = "binary"
	case []*multipart.FileHeader:
		schema = openapi3.NewArraySchema()
		schema.Items = &openapi3.SchemaRef{
			Value: &openapi3.Schema{
				Type:   "string",
				Format: "binary",
			},
		}
	default:

	}
	return schema
}
