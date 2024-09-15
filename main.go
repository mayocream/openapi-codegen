package main

import (
	"fmt"
	"os"
	"path"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	flag "github.com/spf13/pflag"
)

var (
	specFile      = flag.StringP("spec", "i", "openapi.yaml", "Path to OpenAPI spec file")
	outputPath    = flag.StringP("output", "o", ".", "Output path for generated Go file")
	packageName   = flag.StringP("package", "p", "api", "Go package name")
	clientTplFile = flag.StringP("client-tpl", "t", "", "Path to client template file, e.g. client.tmpl")
)

func init() {
	flag.Parse()
}

func main() {
	// Parse the OpenAPI spec
	_, spec, err := parseOpenAPISpec(*specFile)
	if err != nil {
		fmt.Printf("Error parsing OpenAPI spec: %v\n", err)
		return
	}

	// Generate Go code from the spec
	err = generate(spec, *packageName, *outputPath)
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
	} else {
		fmt.Println("Code generated successfully!")
	}
}

// Generate Go code from the OpenAPI spec
func generate(spec *openapi3.T, packageName, outputFilePath string) error {
	code, err := generateComponents(spec, packageName)
	if err != nil {
		return fmt.Errorf("error generating components: %w", err)
	}

	err = os.WriteFile(path.Join(outputFilePath, "schema.gen.go"), []byte(code), 0644)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	// custom client template
	if *clientTplFile != "" {
		fmt.Println("Using custom client template: ", *clientTplFile)
		f, err := os.ReadFile(*clientTplFile)
		if err != nil {
			return fmt.Errorf("error reading client template file: %w", err)
		}
		clientTplContent = string(f)
		clientTpl = template.Must(template.New("client").Funcs(funcMap).Parse(clientTplContent))
	} else {
		fmt.Println("Using default client template: ", "https://github.com/mayocream/openapi-codegen/blob/main/templates/client.tmpl")
	}

	code, err = generateClient(spec, packageName)
	if err != nil {
		return fmt.Errorf("error generating client: %w", err)
	}

	return os.WriteFile(path.Join(outputFilePath, "client.gen.go"), []byte(code), 0644)
}
