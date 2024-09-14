package main

import (
	"flag"
	"fmt"
)

var (
	specFile = flag.String("spec", "openapi.yaml", "Path to OpenAPI spec file")
	outputPath = flag.String("output", ".", "Output path for generated Go file")
	packageName = flag.String("package", "api", "Go package name")
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
	err = generate(spec, *packageName, *outputPath + "/schema.gen.go")
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
	} else {
		fmt.Println("Code generated successfully!")
	}
}
