package fiber_swagger

import (
	"github.com/fatih/structtag"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

func (s *Swagger) addComponents(components ...interface{}) {
	s.Components = append(s.Components, components...)
}

func (s *Swagger) buildComponents() {
	for _, component := range s.Components {
		if component == nil {
			continue
		}
		name, openApiSchemaRef := s.getNameAndOpenApiSchemaRefFromComponent(component)
		s.Schemas[name] = openApiSchemaRef
	}
	s.OpenAPI.Components.Schemas = s.Schemas
}

func (s *Swagger) getNameAndOpenApiSchemaRefFromComponent(component interface{}) (name string, openApiSchemaRef *openapi3.SchemaRef) {
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

	openApiSchemaRef.Value = s.getSchemaFromComponent(component)
	openApiSchemaRef.Value.Title = name
	return name, openApiSchemaRef
}

func (s *Swagger) getSchemaFromComponent(component interface{}) *openapi3.Schema {
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
					s.parseNestedArrayOfStructure(schema, tag.Name, field.Type.Elem().Name())
					s.getSchemaFromComponent(value.Interface())
				}
			}
			if value.Kind() == reflect.Struct {
				v := value.Type().Elem()
				result := isBasicType(v)
				if !result {
					s.parseStructure(schema, tag.Name, value.Interface())
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
		s.parseArrayOfStructure(name, m)
	} else {
		schema = s.getSchemaFromBaseType(component)
	}
	return schema
}

func (s *Swagger) parseArrayOfStructure(name string, model interface{}) {
	result := s.getSchemaFromComponent(model)
	openaiSchemaRef := &openapi3.SchemaRef{
		Value: openapi3.NewAllOfSchema(),
	}
	openaiSchemaRef.Value = result
	openaiSchemaRef.Value.Title = name
	s.Schemas[name] = openaiSchemaRef
}

func (s *Swagger) parseStructure(schema *openapi3.Schema, name string, model interface{}) {
	result := s.getSchemaFromComponent(model)
	openaiSchemaRef := &openapi3.SchemaRef{
		Value: openapi3.NewSchema(),
	}
	openaiSchemaRef.Value = result
	openaiSchemaRef.Value.Title = name
	s.Schemas[name] = openaiSchemaRef
	schemaRef := getSchemaRef(name)
	schema.Properties[name] = schemaRef
}

func (s *Swagger) parseNestedArrayOfStructure(schema *openapi3.Schema, name, fieldName string) {
	openapiSchema := &openapi3.SchemaRef{
		Value: openapi3.NewArraySchema(),
	}
	schemaRef := getSchemaRef(fieldName)
	openapiSchema.Value.Items = schemaRef
	schema.Properties[name] = openapiSchema
}
