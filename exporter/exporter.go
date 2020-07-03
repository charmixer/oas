package exporter

import (
	"fmt"
	"github.com/charmixer/oas/api"
	"gopkg.in/yaml.v2"
	"reflect"
	"regexp"
	"strings"
)

/*
	openapi: 3.0.0
	info:
		title: Sample API
		description: Optional multiline or single-line description in [CommonMark](http://commonmark.org/help/) or HTML.
		version: 0.1.9
	servers:
		- url: http://api.example.com/v1
			description: Optional server description, e.g. Main (production) server
		- url: http://staging-api.example.com
			description: Optional server description, e.g. Internal staging server for testing
	paths:
		/users:
			get:
				summary: Returns a list of users.
				description: Optional extended description in CommonMark or HTML.
				responses:
					'200':    # status code
						description: A JSON array of user names
						content:
							application/json:
								schema:
									type: array
									items:
										type: string
*/

type Item struct {
	Type        string
	Description string
	Properties  interface{} `yaml:",omitempty"`
	Items       interface{}         `yaml:",omitempty"`
}

type Property struct {
	Type                 string
	Description          string
	AdditionalProperties interface{}         `yaml:"additionalProperties,omitempty"`
	Properties           interface{} `yaml:",omitempty"` // nesting
	Items                interface{}         `yaml:",omitempty"`
}

type Request struct {
	Description string
	Content     map[string]struct {
		Schema interface{} // Schema
	}
}

type Response struct {
	Description string
	Content     map[string]struct {
		Schema interface{} // Schema
	}
}

type Path struct {
	Summary     string
	Description string
	Request     Request `yaml:"requestBody,omitempty"`
	Responses   map[int]Response
}

type openapi struct {
	Openapi string
	Info    struct {
		Title       string
		Description string
		Version     string
	}
	Paths map[string]map[string]Path
}

var oasTypeMap = map[string]string{
	"bool":   "boolean",
	"string": "string",
	"slice":  "array",
	"byte": "integer",
	"rune": "integer",
	"int": "integer",
	"int8": "integer",
	"int16": "integer",
	"int32": "integer",
	"int64": "integer",
	"uint": "integer",
	"uint8": "integer",
	"uint16": "integer",
	"uint32": "integer",
	"uint64": "integer",
	"float32": "number",
	"float64": "number",
	"complex64": "number",
	"complex128": "number",
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// convertStructFieldToOas is used to figure out which name to use by reading field tags
func convertStructFieldToOasField(f reflect.StructField) (r string) {
	// FIXME currently hardcoded json
	t := api.CONTENT_TYPE_JSON
	r = f.Name
	switch t {
	case api.CONTENT_TYPE_JSON:
		s := strings.Split(f.Tag.Get("json"), ",")
		if s[0] != "" {
			r = s[0]
		}
		break
	case api.CONTENT_TYPE_XML:
		s := strings.Split(f.Tag.Get("xml"), ",")
		if s[0] != "" {
			r = s[0]
		}
		break
	case api.CONTENT_TYPE_TEXT:
		// undecided
		break
	default:
		break
	}

	return toSnakeCase(r)
}

func goSliceToOas(i interface{}) (interface{}) {
	s := reflect.TypeOf(i)

	elem := s.Elem()

	oas := goToOas(reflect.Zero(elem).Interface())

	item := Item{}
	item.Type = "array"
	item.Items = oas
	return item
}

func goStructToOas(i interface{}) (interface{}) {
	s := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	p := make(map[string]interface{})

	for n := 0; n < s.NumField(); n++ {
		field := s.Field(n)
		value := v.Field(n)

		oasFieldName := convertStructFieldToOasField(field)

		oas := goToOas(value.Interface())
		p[oasFieldName] = oas
	}

	// TODO get description from tags

	return Property{
		Type:        "object",
		Description: "",
		Properties:  p,
	}
}

func goMapToOas(i interface{}) (p Property) {
	s := reflect.TypeOf(i)

	elem := s.Elem()

	oas := goToOas(reflect.Zero(elem).Interface())
	p.Type = "object"
	p.AdditionalProperties = oas

	// TODO get description from tag

	return p
}

func goPrimitiveToOas(k string, i interface{}) Property {
	kind, ok := oasTypeMap[k]

	if !ok {
		panic("Unknown kind " + kind)
	}

	// TODO find description from tags

	return Property{
		Type:        kind,
		Description: "",
	}
}

func goToOas(i interface{}) (r interface{}) {

	t := reflect.TypeOf(i)

	switch t.Kind() {
		/*
		FIXME following types is not handled in any way
		Invalid Kind = iota
    Array
    Chan
    Func
    Interface
    Ptr
    UnsafePointer
		*/
	case reflect.Slice:
		return goSliceToOas(i)
	case reflect.Struct:
		return goStructToOas(i)
	case reflect.Map:
		return goMapToOas(i)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			 reflect.Float32, reflect.Float64,
			 reflect.Complex64, reflect.Complex128,
			 reflect.Bool, reflect.String:
		return goPrimitiveToOas(t.Kind().String(), i)

	default:
		panic("unknown type " + t.Kind().String())
	}

}

func ToOasModel(apiModel api.Api) {
	var oas openapi

	oas.Openapi = "3.0.3"
	oas.Info.Title = apiModel.Title
	oas.Info.Description = apiModel.Description
	oas.Info.Version = apiModel.Version

	oas.Paths = make(map[string]map[string]Path)
	for _, p := range apiModel.GetPaths() {

		path := Path{
			Summary:     p.Summary,
			Description: p.Description,
		}

		if strings.ToLower(p.Method) != "get" {
			path.Request = Request{
				Description: p.Request.Description,
				Content: map[string]struct {
					Schema interface{}
				}{
					api.CONTENT_TYPE_JSON: {Schema: goToOas(p.Request.Schema)},
				},
			}
		}


		responses := make(map[int]Response)
		for _, r := range p.Responses {
			responses[r.Code] = Response{
				Description: r.Description,
				Content: map[string]struct {
					Schema interface{}
				}{
					api.CONTENT_TYPE_JSON: {Schema: goToOas(r.Schema)},
				},
			}
		}
		path.Responses = responses

		oas.Paths[p.Url] = make(map[string]Path)
		oas.Paths[p.Url][strings.ToLower(p.Method)] = path
	}

	d, err := yaml.Marshal(&oas)
	if err != nil {
		fmt.Printf("error: %v", err)
	}
	fmt.Printf("--- m dump:\n%s\n\n", string(d))

}

func ToYaml() {

}
