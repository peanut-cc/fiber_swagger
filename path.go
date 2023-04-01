package fiber_swagger

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/peanut-cc/fiber_swagger/router"
	"net/http"
	"reflect"
	"strings"
)

func (s *Swagger) addPath(route fiber.Route, request interface{}, responses map[int]interface{}) {
	items := strings.Split(route.Path, "/")
	tags := []string{items[3]}
	rt := router.New(route.Path, route.Method, route.Name, tags, router.Request(request), router.Responses(responses))
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
				Responses: s.getResponses(r.Responses),
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

func (s *Swagger) getResponses(responses map[int]interface{}) openapi3.Responses {
	openapiResponse := openapi3.NewResponses()

	for httpCode, rt := range responses {
		switch httpCode {
		case http.StatusOK:
			openapiResponse[fmt.Sprintf("%d", http.StatusOK)] = s.getResponseRef(rt)
		case http.StatusForbidden:
			openapiResponse[fmt.Sprintf("%d", http.StatusForbidden)] = s.ForbiddenResponse()
		case http.StatusNoContent:
			openapiResponse[fmt.Sprintf("%d", http.StatusNoContent)] = s.NoContentResponse()
		case http.StatusInternalServerError:
			openapiResponse[fmt.Sprintf("%d", http.StatusInternalServerError)] = s.getResponseRef(rt)
		case http.StatusBadRequest:
			openapiResponse[fmt.Sprintf("%d", http.StatusBadRequest)] = s.getResponseRef(rt)
		}
	}
	return openapiResponse
}

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
