package main

import (
	"fmt"
	"os"

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

	inSpec, err := spec.NewFromYaml(inStream)
	if err != nil {
		panic(err)
	}

	outSpec := inSpec.Transform()
	err = outSpec.ToYaml(outStream)
	if err != nil {
		panic(err)
	}
}
