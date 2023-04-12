package fiber_swagger

import (
	"github.com/fatih/structtag"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

func (s *Swagger) AddComponents(components ...interface{}) {
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
	if schemaRef, ok := s.Schemas[name]; ok {
		return name, schemaRef
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

			_, err = tags.Get(EMBED)
			if err == nil {
				embedSchema := s.getSchemaFromComponent(value.Interface())
				for key, property := range embedSchema.Properties {
					schema.Properties[key] = property
				}
				schema.Required = append(schema.Required, embedSchema.Required...)
				continue
			}
			var tagName string
			tag, err := tags.Get(JSON)
			if err == nil {
				tagName = tag.Name
				result := isBasicType(value.Type())
				if result {
					fieldSchema = s.getSchemaFromBaseType(value.Interface())
				} else {
					fieldSchema = s.getSchemaFromComponent(value.Interface())
				}
				schema.Properties[tagName] = openapi3.NewSchemaRef("", fieldSchema)
			}
			query, err := tags.Get(QUERY)
			if err == nil {
				tagName = query.Name
				result := isBasicType(value.Type())
				if result {
					fieldSchema = s.getSchemaFromBaseType(value.Interface())
				} else {
					fieldSchema = s.getSchemaFromComponent(value.Interface())
				}
				schema.Properties[tagName] = openapi3.NewSchemaRef("", fieldSchema)
			}

			validateTag, err := tags.Get(VALIDATE)
			if err == nil && validateTag.Name == "required" {
				schema.Required = append(schema.Required, tagName)
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
				var v reflect.Type
				if value.Kind() == reflect.Ptr {
					v = value.Type().Elem()
				} else {
					v = value.Type()
				}

				result := isBasicType(v)
				if !result {
					s.parseStructure(schema, tag.Name, value.Interface())
				}
			}

			descriptionTag, err := tags.Get(DESCRIPTION)
			if err == nil {
				fieldSchema.Description = descriptionTag.Name
			}
		}

	} else if type_.Kind() == reflect.Slice {
		if type_.Elem().Kind() == reflect.Struct {
			name := type_.Elem().Name()
			m := reflect.New(type_.Elem()).Elem().Interface()
			s.parseArrayOfStructure(name, m)
		}
		if isBasicType(type_.Elem()) {
			schema = openapi3.NewArraySchema()
			filedValue := s.getSchemaFromComponent(reflect.New(type_.Elem()).Elem().Interface())
			schema.Items = &openapi3.SchemaRef{Value: filedValue}
		}
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
