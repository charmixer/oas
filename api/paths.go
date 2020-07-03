package api

import (
	"net/http"
)

const (
	CONTENT_TYPE_JSON string = "application/json"
	CONTENT_TYPE_XML  string = "application/xml"
	CONTENT_TYPE_TEXT string = "text/plain"
)

type Api struct {
	// Info Info
	Title string
	Summary string
	Description string
	Version string
	Paths []Path
}

type ParamSchema struct {
	Type   string
	Format string
}

type Param struct {
	Name        string
	Description string
	Required    bool
	Schema      ParamSchema
}

type Request struct {
	Description string
	Required    bool
	Schema      interface{}
}

type Response struct {
	Description string
	Required    bool
	Code        int
	Schema      interface{}
}

type Path struct {
	Summary     string
	Description string
	Url         string
	Method      string
	Handler     http.HandlerFunc
	Path        []Param
	Request     Request
	Responses   []Response
}

func (api *Api) NewPath(method string, url string, handler http.HandlerFunc, path Path) {
	path.Method = method
	path.Url = url
	path.Handler = handler
	api.Paths = append(api.Paths, path)
}

func (api *Api) GetPaths() []Path {
	return api.Paths
}
