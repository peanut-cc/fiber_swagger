package fiber_swagger

import (
	"github.com/invopop/yaml"
	"io"
	"os"
)

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
