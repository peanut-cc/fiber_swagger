package router

type Option func(router *Router)

func Request(request interface{}) Option {
	return func(router *Router) {
		router.Request = request
	}
}

func Responses(response map[int]interface{}) Option {
	return func(router *Router) {
		router.Responses = response
	}
}

func Headers(header interface{}) Option {
	return func(router *Router) {
		router.Header = header
	}
}

func Cookies(cookie interface{}) Option {
	return func(router *Router) {
		router.Cookie = cookie
	}
}

func HasSecurity(hasSecurity bool) Option {
	return func(router *Router) {
		router.HasSecurity = hasSecurity
	}
}
