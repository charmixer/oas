package api

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
	In          string
	Name        string
	Description string
	Required    bool
	ContentType string
	Schema      interface{}
}

type Example struct {
	Summary     string
	Description string
	Schema      interface{}
}

type Request struct {
	Description string
	Required    bool
	ContentType []string
	Schema      interface{}
	Examples    []Example
}

type Response struct {
	Description string
	Required    bool
	Code        int
	ContentType []string
	Schema      interface{}
	Examples    []Example
}

type Path struct {
	Summary     string
	Description string
	Url         string
	Method      string
	Tags        []Tag
	Request     Request
	Responses   []Response
}

type Tag struct {
	Name        string
	Description string
}

func (api *Api) NewEndpoint(method string, url string, path Path) {
	path.Method = method
	path.Url = url
	api.Paths = append(api.Paths, path)
}
