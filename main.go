package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var schemaPaths = []string{
	// paths/{path}/{verb}/requestBody/content/{content-type}/schema
	".paths.*.*.requestBody.content.*.schema",
	// paths/{path}/{verb}/responses/{statusCode}/content/{content-type}/schema
	".paths.*.*.responses.*.content.*.schema",
}

type objectPath struct {
	object Object
	path   Path
}

func (p Path) generateSchemaNameFromRequest() string {
	var sb strings.Builder
	// Write the verb as prefix
	sb.WriteString(strings.ToLower(p[2]))
	sb.WriteString(sanitizeURLPath(p[1]))
	sb.WriteString("Request")
	return sb.String()
}

func (p Path) generateSchemaNameFromResponse() string {
	var sb strings.Builder
	// Write the verb as prefix
	sb.WriteString(strings.ToLower(p[2]))
	sb.WriteString(sanitizeURLPath(p[1]))
	sb.WriteString("Response")
	// Write the response
	sb.WriteString(toTitle(p[6]))
	return sb.String()
}

func sanitizeURLPath(in string) string {
	in = strings.Trim(in, "/")
	vals := strings.Split(in, "/")
	var sb strings.Builder
	for _, v := range vals {
		sb.WriteString(toTitle(v))
	}
	return sb.String()
}

type (
	Object map[interface{}]interface{}
	Spec   struct{ Object }
	Path   []string
)

func NewPath(path string) Path {
	return strings.Split(strings.TrimPrefix(path, "."), ".")
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: openapi-extract-schema {input-file} {output-file}")
		return
	}

	inputFileName := os.Args[1]
	outputFileName := os.Args[2]

	inStream, err := os.Open(inputFileName)
	if err != nil {
		panic(err)
	}

	outStream, err := os.Create(outputFileName)
	if err != nil {
		panic(err)
	}

	spec := Spec{}
	err = yaml.NewDecoder(inStream).Decode(&spec.Object)
	if err != nil {
		panic(err)
	}

	outSpec := spec.Transform()
	err = yaml.NewEncoder(outStream).Encode(outSpec.Object)
	if err != nil {
		panic(err)
	}
}

func (s Spec) Transform() Spec {
	requests := s.FindPath(schemaPaths[0])
	responses := s.FindPath(schemaPaths[1])
	fmt.Printf("Found %d request schema\n", len(requests))
	fmt.Printf("Found %d response schema\n", len(responses))

	for _, val := range requests {
		s.moveToSchemas(val, val.path.generateSchemaNameFromRequest())
	}
	for _, val := range responses {
		s.moveToSchemas(val, val.path.generateSchemaNameFromResponse())
	}

	return s
}

func (s Spec) FindPath(path string) []objectPath {
	path = strings.TrimPrefix(path, ".")
	return s.Object.findPath(NewPath(path), nil)
}

func (o Object) findPath(path Path, parent Path) []objectPath {
	if len(path) == 0 {
		return []objectPath{{object: o, path: parent}}
	}
	switch path[0] {
	case "*":
		ret := []objectPath{}
		for k, v := range o {
			obj, ok := v.(Object)
			if ok {
				key := fmt.Sprintf("%v", k)
				ret = append(ret, obj.findPath(path[1:], append(parent, key))...)
			}
		}
		return ret
	default:
		v, ok := o[path[0]]
		if ok {
			obj, ok := v.(Object)
			if ok {
				return obj.findPath(path[1:], append(parent, path[0]))
			}
		}
		return nil
	}
}

func (s Spec) schemasNode() Object {
	return s.Object.getOrCreateChildObject("components").
		getOrCreateChildObject("schemas")
}

func (o Object) getOrCreateChildObject(name string) Object {
	r, ok := o[name]
	if !ok {
		ret := Object{}
		o[name] = ret
		return ret
	}

	ret, ok := r.(Object)
	if !ok {
		panic(fmt.Errorf("%s is not object", name))
	}
	return ret
}

func (s Spec) moveToSchemas(objPath objectPath, name string) {
	s.schemasNode()[name] = objPath.object
}

func toTitle(in string) string {
	return cases.Title(language.English).String(in)
}
