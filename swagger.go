package fiber_swagger

import (
	"fmt"
	"github.com/fatih/structtag"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/invopop/yaml"
	"io"
	"mime/multipart"
	"os"
	"reflect"
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
	for _, router := range app.GetRoutes() {
		if router.Name == "" {
			continue
		}
		reqRep := s.load(router.Name)
		req := reqRep.Request
		rep := reqRep.Response
		s.addComponent(req)
		s.addComponent(rep)
	}
	s.buildComponents()
	err := s.WriteToYaml()
	if err != nil {
		panic(err)
	}
}

func (s *Swagger) load(name string) (httpResRep *HttpRequestResponse) {
	result, _ := s.routersMap.Load(name)
	return result.(*HttpRequestResponse)
}

func (s *Swagger) addComponent(component interface{}) {
	s.Components = append(s.Components, component)
}

func (s *Swagger) buildComponents() {
	for _, component := range s.Components {
		if component == nil {
			continue
		}
		name, schema := s.getSchemaFromComponent(component)
		s.Schemas[name] = schema
	}
	s.OpenAPI.Components.Schemas = s.Schemas
}

func (s *Swagger) getSchemaFromComponent(component interface{}) (name string, openApiSchemaRef *openapi3.SchemaRef) {
	if component == nil {
		return "", nil
	}
	openApiSchemaRef = &openapi3.SchemaRef{
		Value: openapi3.NewAllOfSchema(),
	}
	if reflect.TypeOf(component).Kind() == reflect.Ptr {
		name = reflect.TypeOf(component).Elem().Name()
	} else {
		name = reflect.TypeOf(component).Name()
	}

	openApiSchemaRef.Value = s.getBodyFromComponent(component)
	return name, openApiSchemaRef
}

func (s *Swagger) getBodyFromComponent(component interface{}) *openapi3.Schema {
	schema := openapi3.NewObjectSchema()
	if component == nil {
		return schema
	}
	type_ := reflect.TypeOf(component)
	value_ := reflect.ValueOf(component)

	if type_.Kind() == reflect.Ptr {
		type_ = type_.Elem()
	}
	if value_.Kind() == reflect.Ptr {
		value_ = value_.Elem()
	}

	if type_.Kind() == reflect.Struct {
		for i := 0; i < type_.NumField(); i++ {
			field := type_.Field(i)
			value := value_.Field(i)
			var fieldSchema *openapi3.Schema

			tags, err := structtag.Parse(string(field.Tag))
			if err != nil {
				panic(err)
			}

			tag, err := tags.Get(JSON)
			if err != nil {
				panic(err)
			} else {
				result := isBasicType(value.Type())
				if result {
					fieldSchema = s.getSchemaFromBaseType(value.Interface())
					schema.Properties[tag.Name] = openapi3.NewSchemaRef("", fieldSchema)
				}

			}

			if value.Kind() == reflect.Slice {
				valueElementType := value.Type().Elem()
				result := isBasicType(valueElementType)
				if !result {
					s.handleNestedStructSlice(schema, tag.Name, field.Type.Elem().Name())
					s.getSchemaFromComponent(value.Interface())
				}
			}
			if value.Kind() == reflect.Struct {
				v := value.Type().Elem()
				result := isBasicType(v)
				if !result {
					s.handleNestedStruct(schema, tag.Name, value.Interface())
				}
			}

			descriptionTag, err := tags.Get(DESCRIPTION)
			if err == nil {
				fieldSchema.Description = descriptionTag.String()
			}

		}

	} else if type_.Kind() == reflect.Slice && type_.Elem().Kind() == reflect.Struct {
		name := type_.Elem().Name()
		m := reflect.New(type_.Elem()).Elem().Interface()
		s.handleStructSlice(name, m)
	} else {
		schema = s.getSchemaFromBaseType(component)
	}
	return schema
}

func (s *Swagger) handleStructSlice(name string, model interface{}) {
	result := s.getBodyFromComponent(model)
	openaiSchemaRef := &openapi3.SchemaRef{
		Value: openapi3.NewAllOfSchema(),
	}
	openaiSchemaRef.Value = result
	openaiSchemaRef.Value.Title = name
	s.Schemas[name] = openaiSchemaRef
}

func (s *Swagger) handleNestedStruct(schema *openapi3.Schema, name string, model interface{}) {
	result := s.getBodyFromComponent(model)
	openaiSchemaRef := &openapi3.SchemaRef{
		Value: openapi3.NewSchema(),
	}
	openaiSchemaRef.Value = result
	openaiSchemaRef.Value.Title = name
	s.Schemas[name] = openaiSchemaRef
	ref := fmt.Sprintf("#/components/schemas/%s", name)
	schemaRef := &openapi3.SchemaRef{Ref: ref}
	schema.Properties[name] = schemaRef
}

func (s *Swagger) handleNestedStructSlice(schema *openapi3.Schema, name, fieldName string) {
	ref := fmt.Sprintf("#/components/schemas/%s", fieldName)
	openapiSchema := &openapi3.SchemaRef{
		Value: openapi3.NewArraySchema(),
	}
	schemaRef := &openapi3.SchemaRef{Ref: ref}
	openapiSchema.Value.Items = schemaRef
	schema.Properties[name] = openapiSchema
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

func (s *Swagger) marshalYAML() ([]byte, error) {
	res, err := s.OpenAPI.MarshalJSON()
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(res)
}

func (s *Swagger) WriteToYaml() error {
	err := ensureDirectory(s.OpenAPIYamlFile)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(s.OpenAPIYamlFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0766)
	if err != nil {
		return err
	}
	content, err := s.marshalYAML()

	if err != nil {
		return err
	}

	if _, err := io.WriteString(file, string(content)); err == nil {
		return nil
	}
	return err
}
