package fiber_swagger

import (
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/peanut-cc/fiber_swagger/router"
	"net/http"
	"reflect"
	"strings"
)

func (s *Swagger) addPath(route fiber.Route, request interface{}, response interface{}) {
	items := strings.Split(route.Path, "/")
	tags := []string{items[3]}
	rt := router.New(route.Path, route.Method, route.Name, tags, router.Request(request), router.Response(response))
	if s.Paths[route.Path] == nil {
		s.Paths[route.Path] = make(map[string]*router.Router)
	}
	s.Paths[route.Path][route.Method] = rt
}

func (s *Swagger) buildPaths() {
	paths := make(openapi3.Paths)
	for path, m := range s.Paths {
		pathItem := &openapi3.PathItem{}
		for method, r := range m {
			operation := &openapi3.Operation{
				Tags:      r.Tags,
				Summary:   r.Description,
				Responses: s.getResponses(r.Response),
			}
			requestBody := s.getRequestBody(r.Request)
			switch method {
			case http.MethodPost:
				pathItem.Post = operation
			case http.MethodPut:
				pathItem.Put = operation
			case http.MethodDelete:
				pathItem.Delete = operation
			}
			if method != http.MethodGet && requestBody.Value.Content != nil {
				operation.RequestBody = requestBody
			}
			operation.Security = &openapi3.SecurityRequirements{}
		}
		paths[path] = pathItem
	}
	s.OpenAPI.Paths = paths
}
func (s *Swagger) getRequestBody(model interface{}) *openapi3.RequestBodyRef {
	body := &openapi3.RequestBodyRef{
		Value: openapi3.NewRequestBody(),
	}
	if model == nil {
		return body
	}
	name := reflect.TypeOf(model).Elem().Name()
	schemaRef := getSchemaRef(s.OpenAPI.Components.Schemas[name].Value.Title)
	body.Value.Required = true
	body.Value.Content = openapi3.Content{
		"application/json": &openapi3.MediaType{Schema: schemaRef},
	}
	return body
}

func (s *Swagger) getResponseRef(response interface{}) *openapi3.ResponseRef {
	name := reflect.TypeOf(response).Elem().Name()
	schemaRef := getSchemaRef(s.OpenAPI.Components.Schemas[name].Value.Title)
	contentMediaType := &openapi3.MediaType{Schema: schemaRef}
	responseRef := &openapi3.ResponseRef{
		Value: &openapi3.Response{
			Description: &SUCCESS,
			Content:     openapi3.Content{"application/json": contentMediaType},
		},
	}
	return responseRef
}

func (s *Swagger) getResponses(response interface{}) openapi3.Responses {
	responses := openapi3.NewResponses()
	responses["200"] = s.getResponseRef(response)
	return responses
}
