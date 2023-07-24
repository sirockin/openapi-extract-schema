package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
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

	spec := map[string]interface{}{}
	err = yaml.NewDecoder(inStream).Decode(&spec)
	if err != nil {
		panic(err)
	}

	outSpec := transform(spec)
	yaml.NewEncoder(outStream).Encode(outSpec)
}

func transform(in map[string]interface{}) map[string]interface{} {
	return in
}
