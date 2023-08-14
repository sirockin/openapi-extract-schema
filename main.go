package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	requestSearchPath             = "$.paths.*.*.requestBody.content.*.schema"
	responseSearchPath            = "$.paths.*.*.responses.*.content.*.schema"
	embeddedObjectSearchPath      = "$.components.schemas.*.properties.*.[?(@type=='object')]"
	embeddedArrayObjectSearchPath = "$.components.schemas.*.*.*.*.[?(@type=='object')]"
)

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
