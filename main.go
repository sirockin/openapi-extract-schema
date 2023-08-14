package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
	"github.com/sirockin/openapi-extract-schema/internal/spec"
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

	inSpec := spec.Spec{}
	err = yaml.NewDecoder(inStream).Decode(&inSpec.Object)
	if err != nil {
		panic(err)
	}

	outSpec := inSpec.Transform()
	err = yaml.NewEncoder(outStream).Encode(outSpec.Object)
	if err != nil {
		panic(err)
	}
}
