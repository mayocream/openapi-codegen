package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"strings"
	"text/template"

	"github.com/samber/lo"
)

//go:embed templates/schema.tmpl
var schemaTplContent string

var schemaTpl = template.Must(template.New("schema").Funcs(template.FuncMap{
	"splitLines": func(s string) []string {
		return strings.Split(s, "\n")
	},
	"pascalCase": lo.PascalCase,
}).Parse(schemaTplContent))

//go:embed templates/client.tmpl
var clientTplContent string

var clientTpl = template.Must(template.New("client").Funcs(template.FuncMap{
	"pascalCase": lo.PascalCase,
}).Parse(clientTplContent))

// applySchemaTemplate Apply Go templates to generate code
func applySchemaTemplate(data any) (string, error) {
	var buf bytes.Buffer
	err := schemaTpl.ExecuteTemplate(&buf, "file", data)
	if err != nil {
		return "", err
	}

	// Format the generated code using go/format
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("error formatting generated code: %w", err)
	}

	return string(formattedCode), nil
}

// applyClientTemplate Apply Go templates to generate code
func applyClientTemplate(data any) (string, error) {
	var buf bytes.Buffer
	err := clientTpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	os.WriteFile("testdata/client.gen.go", buf.Bytes(), 0644)

	// Format the generated code using go/format
	formattedCode, err := format.Source(buf.Bytes())
	if err != nil {
		return "", fmt.Errorf("error formatting generated code: %w", err)
	}

	return string(formattedCode), nil
}
