package router

type Router struct {
	Path        string
	Method      string
	Description string
	Deprecated  bool
	Request     interface{}
	Response    interface{}
	Header      interface{}
	Cookie      interface{}
	Tags        []string
	HasSecurity bool
}

func New(path, method, description string, tags []string, options ...Option) *Router {
	r := &Router{
		Path:        path,
		Method:      method,
		Description: description,
		Tags:        tags,
	}

	for _, option := range options {
		option(r)
	}

	return r
}
