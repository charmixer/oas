package exporter

import (
	"github.com/charmixer/oas/api"
	//"gopkg.in/yaml.v2"
	yaml "github.com/ghodss/yaml"
	"encoding/json"
	"reflect"
	"strings"
	"fmt"
)

type Item struct {
	Type        string `yaml:"type" json:"type"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Properties  interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	Items       interface{} `yaml:"items,omitempty" json:"items,omitempty"`
	//Example	  	interface{} `yaml:",omitempty" json:",omitempty"`
}

type Property struct {
	Type                 string `yaml:"type" json:"type"`
	Description          string `yaml:"description,omitempty" json:"description,omitempty"`
	AdditionalProperties interface{} `yaml:"additionalProperties,omitempty" json:"additionalProperties,omitempty"`
	Properties  interface{} `yaml:"properties,omitempty" json:"properties,omitempty"`
	Items       interface{} `yaml:"items,omitempty" json:"items,omitempty"`
	//Example							 interface{} `yaml:",omitempty" json:",omitempty"`
}

type Example struct {
	Summary string `yaml:"summary,omitempty" json:"summary,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Value interface{} `yaml:"value,omitempty" json:"value,omitempty"`
}

type Content map[string]struct {
	Schema interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`
	Example interface{} `yaml:"example,omitempty" json:"example,omitempty"`
	Examples map[string]Example `yaml:"examples,omitempty" json:"examples,omitempty"`
}

type Request struct {
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Content     Content `yaml:"content,omitempty" json:"content,omitempty"`
}

type Response struct {
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Content     Content `yaml:"content,omitempty" json:"content,omitempty"`
}

type Param struct {
	In					string `yaml:"in" json:"in"`
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Required    bool `yaml:"required" json:"required"`
	Content     Content `yaml:"content,omitempty" json:"content,omitempty"`
	Schema interface{} `yaml:"schema,omitempty" json:"schema,omitempty"`
}

type Tag struct {
	Name string `yaml:"name" json:"name"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

type Path struct {
	Summary     string `yaml:"summary" json:"summary"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Tags			  []string `yaml:"tags" json:"tags"`
	Parameters  []Param `yaml:"parameters,omitempty" json:"parameters,omitempty"`
	Request     *Request `yaml:"requestBody,omitempty" json:"requestBody,omitempty"`
	Responses   map[int]Response `yaml:"responses" json:"responses"`
}

type Openapi struct {
	Openapi string `yaml:"openapi" json:"openapi"`
	Info    struct {
		Title       string `yaml:"title" json:"title"`
		Description string `yaml:"description,omitempty" json:"description,omitempty"`
		Version     string `yaml:"version" json:"version"`
	} `yaml:"info" json:"info"`
	Paths map[string]map[string]Path `yaml:"paths" json:"paths"`
	Tags []Tag `yaml:"tags" json:"tags"`
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

// convertStructFieldToOas is used to figure out which name to use by reading field tags
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

func goSliceToOas(i interface{}) interface{} {
	s := reflect.TypeOf(i)

	elem := s.Elem()

	oas, _ := goToOas(reflect.Zero(elem).Interface())

	item := Item{}
	item.Type = "array"
	item.Items = oas
	return item
}

func goStructToOas(i interface{}) interface{} {
	s := reflect.TypeOf(i)
	v := reflect.ValueOf(i)

	p := make(map[string]interface{})

	for n := 0; n < s.NumField(); n++ {
		field := s.Field(n)
		value := v.Field(n)

		oasFieldName := convertStructFieldToOasField(field)

		oas, _ := goToOas(value.Interface())
		switch t := oas.(type) {
			case Property:
				prop := oas.(Property)
				prop.Description = field.Tag.Get("oas")
				oas = prop
			case Item:
				item := oas.(Item)
				item.Description = field.Tag.Get("oas")
				oas = item
			case nil:
				oas = Property{
					Type: "object",
					AdditionalProperties: true,
				}
			default:
				var r = reflect.TypeOf(t)
				panic(fmt.Sprintf("Unhandled type '%v' from goToOas func", r))
		}
		p[oasFieldName] = oas
	}

	// TODO get description from tags - how to describe outer struct?

	return Property{
		Type:        "object",
		Description: s.Name() + " object",
		Properties:  p,
	}
}

func goMapToOas(i interface{}) (p Property) {
	s := reflect.TypeOf(i)

	elem := s.Elem()

	oas, _ := goToOas(reflect.Zero(elem).Interface())
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

func goToOas(i interface{}) (r interface{}, kind reflect.Kind) {

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
		return goSliceToOas(i), t.Kind()
	case reflect.Struct:
		return goStructToOas(i), t.Kind()
	case reflect.Map:
		return goMapToOas(i), t.Kind()

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.Bool, reflect.String:
		return goPrimitiveToOas(t.Kind().String(), i), t.Kind()

	default:
		panic("unknown type " + t.Kind().String())
	}

}

func goToOasParameters(i interface{}) (params []Param) {
	typeOf := reflect.TypeOf(i)
	valueOf := reflect.ValueOf(i)

	if typeOf == nil {
		return params
	}

	if typeOf.Kind() != reflect.Struct {
		panic("Only a struct can be parsed into parameters, "+typeOf.Kind().String()+" ("+typeOf.Name()+") given")
	}

	for i := 0; i < valueOf.NumField(); i++ {
		f := valueOf.Field(i)
		t := typeOf.Field(i)

		tags := make(map[string]string, 3)

		query := t.Tag.Get("query")
		if query != "" && query != "-" {
			tags["query"] = query
		}

		header := t.Tag.Get("header")
		if header != "" && header != "-" {
			tags["header"] = header
		}

		cookie := t.Tag.Get("cookie")
		if cookie != "" && cookie != "-" {
			tags["cookie"] = cookie
		}

		description := t.Tag.Get("oas")

		for in, name := range tags {
			param := Param{
				In: in,
				Name: name,
				Description: description,
			}

			schema, kind := goToOas(f.Interface())

			// controlled by result of gotooas
			if kind == reflect.Struct || kind == reflect.Map {
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

func ToOasModel(apiModel api.Api) (oas Openapi) {
	oas.Openapi = "3.0.3"
	oas.Info.Title = apiModel.Title
	oas.Info.Description = apiModel.Description
	oas.Info.Version = apiModel.Version

	var tags = make(map[string]Tag)

	oas.Paths = make(map[string]map[string]Path)
	for _, p := range apiModel.GetPaths() {

		path := Path{
			Summary:     p.Summary,
			Description: p.Description,
		}

		if strings.ToLower(p.Method) != "get" {
			contentType := api.CONTENT_TYPE_JSON
			if p.Request.ContentType != "" {
				contentType = p.Request.ContentType
			}

			schema, _ := goToOas(p.Request.Schema)

			examples := make(map[string]Example)
			for _, e := range p.Request.Examples {
				example := Example{
					Summary: e.Summary,
					Description: e.Description,
					Value: e.Schema,
				}
				examples[fmt.Sprintf("Sample %d", len(examples) + 1)] = example
			}

			path.Request = &Request{
				Description: p.Request.Description,
				Content: Content{
					contentType: {
						Schema: schema,
						Examples: examples,
					},
				},
			}
		}

		path.Parameters = goToOasParameters(p.Request.Params)

		responses := make(map[int]Response)
		for _, r := range p.Responses {
			contentType := api.CONTENT_TYPE_JSON
			if r.ContentType != "" {
				contentType = r.ContentType
			}

			schema, _ := goToOas(r.Schema)

			examples := make(map[string]Example)
			for _, e := range r.Examples {
				example := Example{
					Summary: e.Summary,
					Description: e.Description,
					Value: e.Schema,
				}
				examples[fmt.Sprintf("Sample %d", len(examples) + 1)] = example
			}

			responses[r.Code] = Response{
				Description: r.Description,
				Content: Content{
					contentType: {
						Schema: schema,
						Examples: examples,
					},
				},
			}
		}
		path.Responses = responses

		for _, t := range p.Tags {
			tags[t.Name] = Tag{
				Name: t.Name,
				Description: t.Description,
			}
			path.Tags = append(path.Tags, t.Name)
		}

		if _, ok := oas.Paths[p.Url]; !ok {
			oas.Paths[p.Url] = make(map[string]Path)
		}
		oas.Paths[p.Url][strings.ToLower(p.Method)] = path
	}

	for _, tag := range tags {
		oas.Tags = append(oas.Tags, Tag{
			Name: tag.Name,
			Description: tag.Description,
		})
	}

	return oas
}


func ToYaml(oas Openapi) (string, error){
	bytes, err := yaml.Marshal(oas)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ToJson(oas Openapi) (string, error){
	d, err := json.Marshal(&oas)
	if err != nil {
		return "", err
	}
	return string(d), nil
}
