package fiber_swagger

import (
	"github.com/fatih/structtag"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

func (s *Swagger) parseParameters(parameters ...interface{}) openapi3.Parameters {
	openapiParameters := openapi3.NewParameters()
	for _, parameter := range parameters {
		if parameter == nil {
			continue
		}
		type_ := reflect.TypeOf(parameter)
		value_ := reflect.ValueOf(parameter)
		if type_.Kind() == reflect.Ptr {
			type_ = type_.Elem()
		}
		if value_.Kind() == reflect.Ptr {
			value_ = value_.Elem()
		}
		if type_.Kind() == reflect.Struct {
			for i := 0; i < type_.NumField(); i++ {
				if type_.Kind() == reflect.Struct && value_.Kind() == reflect.Invalid {
					value_ = reflect.New(type_).Elem()
				}
				field := type_.Field(i)
				value := value_.Field(i)
				tags, err := structtag.Parse(string(field.Tag))
				if err != nil {
					panic(err)
				}
				_, err = tags.Get(EMBED)
				if err == nil {
					embedParameters := s.parseParameters(value.Interface())
					parameters = append(parameters, embedParameters)
				}
				p := &openapi3.Parameter{
					Schema: openapi3.NewSchemaRef("", s.getSchemaFromBaseType(value.Interface())),
				}
				jsonTag, err := tags.Get(JSON)
				if err == nil {
					p.In = "json"
					p.Name = jsonTag.Name
				}
				queryTag, err := tags.Get(QUERY)
				if err == nil {
					p.In = openapi3.ParameterInQuery
					p.Name = queryTag.Name
				}
				uriTag, err := tags.Get(URI)
				if err == nil {
					p.In = openapi3.ParameterInPath
					p.Name = uriTag.Name
				}
				headerTag, err := tags.Get(HEADER)
				if err == nil {
					p.In = openapi3.ParameterInHeader
					p.Name = headerTag.Name
				}
				cookieTag, err := tags.Get(COOKIE)
				if err == nil {
					p.In = openapi3.ParameterInCookie
					p.Name = cookieTag.Name
				}
				if p.In == "" {
					continue
				}
				descriptionTag, err := tags.Get(DESCRIPTION)
				if err == nil {
					p.WithDescription(descriptionTag.Name)
				}
				validateTag, err := tags.Get(VALIDATE)
				if err == nil {
					p.WithRequired(validateTag.Name == "required")
				}
				defaultTag, err := tags.Get(DEFAULT)
				if err == nil {
					p.Schema.Value.WithDefault(defaultTag.Name)
				}
				exampleTag, err := tags.Get(EXAMPLE)
				if err == nil {
					p.Schema.Value.Example = exampleTag.Name
				}
				openapiParameters = append(openapiParameters, &openapi3.ParameterRef{
					Value: p,
				})
			}
		}

	}
	return openapiParameters
}
