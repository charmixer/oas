package api

import (
	"net/http"
)

const (
	CONTENT_TYPE_JSON string = "application/json"
	CONTENT_TYPE_XML  string = "application/xml"
	CONTENT_TYPE_TEXT string = "text/plain"
	CONTENT_TYPE_HTML string = "text/html"
)

type Api struct {
	// Info Info
	Title       string
	Summary     string
	Description string
	Version     string
	Paths       []Path
}

type Param struct {
	In					string
	Name        string
	Description string
	Required    bool
	ContentType	string
	Schema      interface{}
}

type Example struct {
	Summary string
	Description string
	Schema interface{}
}

type Request struct {
	Description string
	Required    bool
	ContentType	string
	Schema      interface{}
/*	Query				interface{}
	Header			interface{}
	Cookie			interface{}*/
	Params      interface{}
	Examples		[]Example
}

type Response struct {
	Description string
	Required    bool
	Code        int
	ContentType	string
	Schema      interface{}
	Examples		[]Example
}

type Path struct {
	Summary     string
	Description string
	Url         string
	Method      string
	Tags				[]Tag
	Handler     http.HandlerFunc
	//Params      []Param
	Request     Request
	Responses   []Response
}

type Tag struct {
	Name string
	Description string
}

func (api *Api) NewPath(method string, url string, handler http.HandlerFunc, path Path, tags []Tag) {
	path.Tags = tags
	path.Method = method
	path.Url = url
	path.Handler = handler
	api.Paths = append(api.Paths, path)
}

func (api *Api) GetPaths() []Path {
	return api.Paths
}
