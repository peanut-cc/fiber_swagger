package fiber_swagger

import "github.com/getkin/kin-openapi/openapi3"

type Option func(swagger *Swagger)

func Title(title string) Option {
	return func(swagger *Swagger) {
		swagger.Title = title
	}
}

func Description(description string) Option {
	return func(swagger *Swagger) {
		swagger.Description = description
	}
}

func Version(version string) Option {
	return func(swagger *Swagger) {
		swagger.Version = version
	}
}

func Contact(name, url, email string) Option {
	contact := &openapi3.Contact{
		Name:  name,
		URL:   url,
		Email: email,
	}
	return func(swagger *Swagger) {
		swagger.Contact = contact
	}
}

func License(name, url string) Option {
	license := &openapi3.License{
		Name: name,
		URL:  url,
	}
	return func(swagger *Swagger) {
		swagger.License = license
	}
}

func DocPath(path string) Option {
	if path == "" {
		path = "./docs/openapi.yaml"
	}
	return func(swagger *Swagger) {
		swagger.OpenAPIYamlFile = path
	}
}
