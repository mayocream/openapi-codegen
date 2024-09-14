package main

import (
	"fmt"

	flag "github.com/spf13/pflag"
)

var (
	specFile    = flag.StringP("spec", "i", "openapi.yaml", "Path to OpenAPI spec file")
	outputPath  = flag.StringP("output", "o", ".", "Output path for generated Go file")
	packageName = flag.StringP("package", "p", "api", "Go package name")
)

func init() {
	flag.Parse()
}

func main() {
	// Parse the OpenAPI spec
	spec, err := parseOpenAPISpec(*specFile)
	if err != nil {
		fmt.Printf("Error parsing OpenAPI spec: %v\n", err)
		return
	}

	// Generate Go code from the spec
	err = generate(spec, *packageName, *outputPath+"/schema.gen.go")
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
	} else {
		fmt.Println("Code generated successfully!")
	}
}
