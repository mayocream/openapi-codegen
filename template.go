package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"strings"
	"text/template"
)

//go:embed templates/schema.tpl
var schemaTplContent string

var (
	schemaTpl = template.Must(template.New("schema").Funcs(template.FuncMap{
		"splitLines": func(s string) []string {
			return strings.Split(s, "\n")
		},
	}).Parse(schemaTplContent))
)

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