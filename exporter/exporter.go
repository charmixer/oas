package exporter

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/charmixer/oas/api"
)

type Item struct {
	Type        string                 `yaml:"type" json:"type"`
	Description string                 `yaml:"description,omitempty" json:"description,omitempty"`
	Properties  map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	Items       interface{}            `yaml:"items,omitempty" json:"items,omitempty"`
	//Example	  	interface{} `yaml:",omitempty" json:",omitempty"`
}

type Property struct {
	Type                 string                 `yaml:"type" json:"type"`
	Description          string                 `yaml:"description,omitempty" json:"description,omitempty"`
	AdditionalProperties interface{}            `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	Properties           map[string]interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	Items                interface{}            `yaml:"items,omitempty" json:"items,omitempty"`
	//Example							 interface{} `yaml:",omitempty" json:",omitempty"`
}

type Example struct {
	Summary     string      `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string      `yaml:"description,omitempty" json:"description,omitempty"`
	Value       interface{} `yaml:"value,omitempty" json:"value,omitempty"`
}

type ContentBody struct {
	Schema   interface{}        `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example  interface{}        `yaml:"example,omitempty" json:"example,omitempty"`
	Examples map[string]Example `yaml:"examples,omitempty" json:"examples,omitempty"`
}
type Content map[string]ContentBody

type Request struct {
	Description string  `yaml:"description" json:"description"`
	Content     Content `yaml:"content,omitempty" json:"content,omitempty"`
}

type HeaderSchema struct {
	Type   string `yaml:"type" json:"type"`
	Format string `yaml:"format,omitempty" json:"format,omitempty"`
}

type ResponseHeader struct {
	Schema      HeaderSchema `yaml:"schema,omitempty" json:"schema,omitempty"`
	Description string       `yaml:"description" json:"description"`
}
type Response struct {
	Description string                    `yaml:"description" json:"description"`
	Headers     map[string]ResponseHeader `yaml:"headers,omitempty" json:"headers,omitempty"`
	Content     Content                   `yaml:"content,omitempty" json:"content,omitempty"`
}

type Param struct {
	In          string      `yaml:"in" json:"in"`
	Name        string      `yaml:"name" json:"name"`
	Description string      `yaml:"description" json:"description"`
	Required    bool        `yaml:"required" json:"required"`
	Content     Content     `yaml:"content,omitempty" json:"content,omitempty"`
	Schema      interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`
}

type Tag struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type Path struct {
	Summary     string           `yaml:"summary" json:"summary"`
	Description string           `yaml:"description" json:"description"`
	Tags        []string         `yaml:"tags" json:"tags"`
	Parameters  []Param          `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Request     Request          `yaml:"requestBody,omitempty" json:"requestBody,omitempty"`
	Responses   map[int]Response `yaml:"responses" json:"responses"`
}

type oasConfig struct {
	tagQuery       string
	tagHeader      string
	tagCookie      string
	tagDescription string
}

type Openapi struct {
	config  oasConfig
	Openapi string `yaml:"openapi" json:"openapi"`
	Info    struct {
		Title       string `yaml:"title" json:"title"`
		Description string `yaml:"description" json:"description"`
		Version     string `yaml:"version" json:"version"`
	} `yaml:"info" json:"info"`
	Paths map[string]map[string]Path `yaml:"paths" json:"paths"`
	Tags  []Tag                      `yaml:"tags" json:"tags"`
}

type OasOption func(e *Openapi)

func WithQueryTag(tag string) OasOption {
	return func(oas *Openapi) {
		oas.config.tagQuery = tag
	}
}
func WithHeaderTag(tag string) OasOption {
	return func(oas *Openapi) {
		oas.config.tagHeader = tag
	}
}
func WithCookieTag(tag string) OasOption {
	return func(oas *Openapi) {
		oas.config.tagCookie = tag
	}
}
func WithDescriptionTag(tag string) OasOption {
	return func(oas *Openapi) {
		oas.config.tagDescription = tag
	}
}

var oasTypeMap = map[string]string{
	"bool":       "boolean",
	"string":     "string",
	"slice":      "array",
	"byte":       "integer",
	"rune":       "integer",
	"int":        "integer",
	"int8":       "integer",
	"int16":      "integer",
	"int32":      "integer",
	"int64":      "integer",
	"uint":       "integer",
	"uint8":      "integer",
	"uint16":     "integer",
	"uint32":     "integer",
	"uint64":     "integer",
	"float32":    "number",
	"float64":    "number",
	"complex64":  "number",
	"complex128": "number",
}

// Used to figure out which name to use by reading field tags
func convertStructFieldToOasField(f reflect.StructField) (r string) {
	t := api.CONTENT_TYPE_JSON // FIXME currently hardcoded json
	r = f.Name
	switch t {
	case api.CONTENT_TYPE_JSON:
		s := strings.Split(f.Tag.Get("json"), ",")
		if s[0] != "" {
			r = s[0]
		}
		break
	default:
		break
	}

	return r
}

func (openapi *Openapi) goSliceToOas(i interface{}) interface{} {
	s := reflect.TypeOf(i)

	elem := s.Elem()

	oas, _ := openapi.goToOas(reflect.Zero(elem).Interface())

	return Item{
		Type:  "array",
		Items: oas,
	}
}

func (openapi *Openapi) goStructToOas(i interface{}) interface{} {
	s := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	p := make(map[string]interface{})

	for n := 0; n < s.NumField(); n++ {
		field := s.Field(n)
		value := v.Field(n)

		// Unexported - skip, note IsExported is ^go1.17
		if !value.CanInterface() {
			continue
		}

		// skip field if found in params
		query := field.Tag.Get(openapi.config.tagQuery)
		header := field.Tag.Get(openapi.config.tagHeader)
		cookie := field.Tag.Get(openapi.config.tagCookie)
		if query != "" || header != "" || cookie != "" {
			continue
		}

		// Flatten embedded struct to (makes embedded struct look in docs like its the same)
		if value.Kind() == reflect.Struct && field.Anonymous {
			oasStruct := openapi.goStructToOas(value.Interface()).(Property)
			for name, oasStructProperties := range oasStruct.Properties {
				p[name] = oasStructProperties
			}
			continue
		}

		oas, _ := openapi.goToOas(value.Interface())
		switch t := oas.(type) {
		case Property:
			prop := oas.(Property)
			prop.Description = field.Tag.Get(openapi.config.tagDescription)
			oas = prop
		case Item:
			item := oas.(Item)
			item.Description = field.Tag.Get(openapi.config.tagDescription)
			oas = item
		case nil:
			oas = Property{
				Type:                 "object",
				AdditionalProperties: true,
			}
		default:
			panic(fmt.Sprintf("Unhandled type '%v' from goToOas func", reflect.TypeOf(t)))
		}
		oasFieldName := convertStructFieldToOasField(field)
		p[oasFieldName] = oas
	}

	// TODO get description from tags - how to describe outer struct?

	return Property{
		Type:        "object",
		Description: s.Name() + " object",
		Properties:  p,
	}
}

func (openapi *Openapi) goMapToOas(i interface{}) (p Property) {
	s := reflect.TypeOf(i)

	elem := s.Elem()

	oas, _ := openapi.goToOas(reflect.Zero(elem).Interface())
	p.Type = "object"
	p.AdditionalProperties = oas

	// TODO get description from tag

	return p
}

func (openapi *Openapi) goPrimitiveToOas(k string, i interface{}) Property {
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

func (openapi *Openapi) goToOas(i interface{}) (r interface{}, kind reflect.Kind) {

	t := reflect.TypeOf(i)

	if t == nil {
		return r, kind
	}

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
		return openapi.goSliceToOas(i), t.Kind()
	case reflect.Struct:
		return openapi.goStructToOas(i), t.Kind()
	case reflect.Map:
		return openapi.goMapToOas(i), t.Kind()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.Bool, reflect.String:
		return openapi.goPrimitiveToOas(t.Kind().String(), i), t.Kind()

	default:
		panic("unknown type " + t.Kind().String())
	}

}

func (openapi *Openapi) goToOasResponseHeaders(i interface{}) (headers map[string]ResponseHeader) {
	typeOf := reflect.TypeOf(i)
	valueOf := reflect.ValueOf(i)

	if typeOf == nil {
		return headers
	}

	if typeOf.Kind() != reflect.Struct {
		panic("Only a struct can be parsed into response headers, " + typeOf.Kind().String() + " (" + typeOf.Name() + ") given")
	}

	headers = make(map[string]ResponseHeader)
	for i := 0; i < valueOf.NumField(); i++ {
		f := valueOf.Field(i)
		t := typeOf.Field(i)

		header := t.Tag.Get(openapi.config.tagHeader)
		if header == "" || header == "-" {
			continue
		}

		description := t.Tag.Get(openapi.config.tagDescription)

		responseHeader := ResponseHeader{
			Description: description,
		}

		schema, _ := openapi.goToOas(f.Interface())

		if prop, ok := schema.(Property); ok {
			responseHeader.Schema = HeaderSchema{
				Type: prop.Type,
			}
		} else {
			panic("Expected property from response header reflection")
		}

		headers[header] = responseHeader
	}

	return headers
}

func (openapi *Openapi) goToOasParameters(i interface{}) (params []Param) {
	typeOf := reflect.TypeOf(i)
	valueOf := reflect.ValueOf(i)

	if typeOf == nil {
		return params
	}

	if typeOf.Kind() != reflect.Struct {
		panic("Only a struct can be parsed into parameters, " + typeOf.Kind().String() + " (" + typeOf.Name() + ") given")
	}

	for i := 0; i < valueOf.NumField(); i++ {
		f := valueOf.Field(i)
		t := typeOf.Field(i)

		tags := make(map[string]string, 3)

		query := t.Tag.Get(openapi.config.tagQuery)
		if query != "" && query != "-" {
			tags["query"] = query
		}

		header := t.Tag.Get(openapi.config.tagHeader)
		if header != "" && header != "-" {
			tags["header"] = header
		}

		cookie := t.Tag.Get(openapi.config.tagCookie)
		if cookie != "" && cookie != "-" {
			tags["cookie"] = cookie
		}

		description := t.Tag.Get(openapi.config.tagDescription)

		for in, name := range tags {
			param := Param{
				In:          in,
				Name:        name,
				Description: description,
			}

			schema, kind := openapi.goToOas(f.Interface())

			// controlled by result of gotooas
			if kind == reflect.Struct || kind == reflect.Map {
				// Describes complex datastructures for parameters like ?filter={"type":"t-shirt","color":"blue"}
				param.Content = Content{
					"application/json": {Schema: schema}, // FIXME content type
				}
			} else {
				param.Schema = schema
			}

			params = append(params, param)
		}
	}

	return params
}

func ToOasModel(apiModel api.Api, options ...OasOption) (openapi Openapi) {

	openapi.config = oasConfig{
		tagQuery:       "oas-query",
		tagHeader:      "oas-header",
		tagCookie:      "oas-cookie",
		tagDescription: "oas-desc",
	}

	for _, opt := range options {
		opt(&openapi)
	}

	openapi.Openapi = "3.0.3"
	openapi.Info.Title = apiModel.Title
	openapi.Info.Description = apiModel.Description
	openapi.Info.Version = apiModel.Version

	var tags = make(map[string]Tag)

	openapi.Paths = make(map[string]map[string]Path)
	for _, p := range apiModel.Paths {

		path := Path{
			Summary:     p.Summary,
			Description: p.Description,
		}

		if strings.ToLower(p.Method) != "get" {
			schema, _ := openapi.goToOas(p.Request.Schema)

			examples := make(map[string]Example)
			for _, e := range p.Request.Examples {
				example := Example{
					Summary:     e.Summary,
					Description: e.Description,
					Value:       e.Schema,
				}
				examples[fmt.Sprintf("Sample %d", len(examples)+1)] = example
			}

			if len(p.Request.ContentType) <= 0 {
				p.Request.ContentType = []string{api.CONTENT_TYPE_JSON}
			}

			emptyProperty := false
			if prop, ok := schema.(Property); ok {
				if len(prop.Properties) == 0 {
					emptyProperty = true
				}
			}

			var content map[string]ContentBody
			if !emptyProperty {
				content := map[string]ContentBody{}
				for _, c := range p.Request.ContentType {
					content[c] = ContentBody{
						Schema:   schema,
						Examples: examples,
					}
				}
			}

			path.Request = Request{
				Description: p.Request.Description,
				Content:     content,
			}
		}

		path.Parameters = openapi.goToOasParameters(p.Request.Schema)

		responses := make(map[int]Response)
		for _, r := range p.Responses {
			schema, _ := openapi.goToOas(r.Schema)

			examples := make(map[string]Example)
			for _, e := range r.Examples {
				example := Example{
					Summary:     e.Summary,
					Description: e.Description,
					Value:       e.Schema,
				}
				examples[fmt.Sprintf("Sample %d", len(examples)+1)] = example
			}

			if len(r.ContentType) <= 0 {
				r.ContentType = []string{api.CONTENT_TYPE_JSON}
			}

			emptyProperty := false
			if prop, ok := schema.(Property); ok {
				if len(prop.Properties) == 0 {
					emptyProperty = true
				}
			}

			var content map[string]ContentBody
			if !emptyProperty {
				content = map[string]ContentBody{}
				for _, c := range r.ContentType {
					content[c] = ContentBody{
						Schema:   schema,
						Examples: examples,
					}
				}
			}

			responses[r.Code] = Response{
				Description: r.Description,
				Headers:     openapi.goToOasResponseHeaders(r.Schema),
				Content:     content,
			}
		}
		path.Responses = responses

		for _, t := range p.Tags {
			tags[t.Name] = Tag{
				Name:        t.Name,
				Description: t.Description,
			}
			path.Tags = append(path.Tags, t.Name)
		}

		if _, ok := openapi.Paths[p.Url]; !ok {
			openapi.Paths[p.Url] = make(map[string]Path)
		}
		openapi.Paths[p.Url][strings.ToLower(p.Method)] = path
	}

	for _, tag := range tags {
		openapi.Tags = append(openapi.Tags, Tag{
			Name:        tag.Name,
			Description: tag.Description,
		})
	}

	return openapi
}
