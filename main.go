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

type Object = map[interface{}]interface{}

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
	fmt.Printf("Found %d request schema\n", len(requests) )
	fmt.Printf("Found %d response schema\n", len(responses) )
	return in
}


func FindPath(path string, spec Object) []Object {
	path = strings.TrimPrefix(path, ".")
	return _findPath(strings.Split(path, "."), spec)
}

func _findPath(path []string, val Object) []Object {
	if len(path)==0{
		return []Object{val}
	}
	switch path[0] {
	case "*":
		ret := []Object{}
		for _, v := range val {
			obj, ok := v.(Object)
			if ok {
				ret = append(ret, _findPath(path[1:], obj)...)
			}
		}
		return ret
	default:
		v, ok := val[path[0]]
		if ok {
			obj, ok := v.(Object)
			if ok {
				return _findPath(path[1:], obj)
			}
		}
		return nil
	}
}
