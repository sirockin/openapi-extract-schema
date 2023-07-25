package main

import (
	"fmt"
	"os"
	"strings"

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

type Object = map[interface{}]interface{}
type Path = []string

func NewPath(path string) Path{
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

	spec := Object{}
	err = yaml.NewDecoder(inStream).Decode(&spec)
	if err != nil {
		panic(err)
	}

	outSpec := transform(spec)
	yaml.NewEncoder(outStream).Encode(outSpec)
}

func transform(in Object) Object {
	requests := FindPath(schemaPaths[0], in)
	responses := FindPath(schemaPaths[1], in)
	fmt.Printf("Found %d request schema\n", len(requests))
	fmt.Printf("Found %d response schema\n", len(responses))
	return in
}

func FindPath(path string, spec Object) []objectPath {
	path = strings.TrimPrefix(path, ".")
	return _findPath(NewPath(path), spec, nil)
}

func _findPath(path Path, val Object, parent Path) []objectPath {
	if len(path) == 0 {
		return []objectPath{ { object:val, path:parent } }
	}
	switch path[0] {
	case "*":
		ret := []objectPath{}
		for k, v := range val {
			obj, ok := v.(Object)
			if ok {
				key := fmt.Sprintf("%v", k)
				ret = append(ret, _findPath(path[1:], obj, append(parent, key))...)
			}
		}
		return ret
	default:
		v, ok := val[path[0]]
		if ok {
			obj, ok := v.(Object)
			if ok {
				return _findPath(path[1:], obj, append(parent, path[0]))
			}
		}
		return nil
	}
}
