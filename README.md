# openapi-codegen

Implementation of OpenAPI code generation in Go.

Supports [OpenAPI 3.0.3](https://spec.openapis.org/oas/v3.0.3.html).

## Features

- [x] Type-safe structs generation based on [kin-openapi](https://github.com/getkin/kin-openapi) package
- [x] HTTP Client generation based on [resty](https://github.com/go-resty/resty) package

## Usage

```bash
go install github.com/mayocream/openapi-codegen@latest

Usage of openapi-codegen:
  -t, --client-tpl string   Path to client template file, e.g. client.tmpl
  -o, --output string       Output path for generated Go file (default ".")
  -p, --package string      Go package name (default "api")
  -i, --spec string         Path to OpenAPI spec file (default "openapi.yaml")

openapi-codegen -i openapi.yaml -o . -p api
```

## Customization

You can customize the generated code by providing a custom template file. The template file should be a Go template file with the following functions:
- pascalCase
- goComment

Example template file:
```go
package {{ .PackageName }}

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/samber/lo"
)

type Client struct {
	client  *resty.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		client:  resty.New().SetBaseURL(baseURL),
	}
}

{{ range .Requests }}
{{- if .Parameters }}
// {{ .Name }}Params represents the parameters for the {{ .Name }} request
type {{ .Name }}Params struct {
	{{- range .Parameters }}
	{{- if .Description }} {{ goComment .Name .Description }} {{- end }}
	{{- if eq .Type "enum" }}
	// {{ .Name }} enum
	{{ .Name }} string `json:"{{ .JSONName }}{{ if .OmitEmpty }},omitempty{{ end }}"`
	{{- else }}
	{{ .Name }} {{ .Type }} `json:"{{ .JSONName }}{{ if .OmitEmpty }},omitempty{{ end }}"`
	{{- end }}
	{{- end }}
}
{{- end }}
func (c *Client) {{ .Name }}({{ if .Parameters }}params {{ .Name }}Params{{ end }}{{ if .Body }}, body {{ .Body.Name }} {{ end }}) ({{ if .Response }}*{{ .Response.Name }}, {{ end }}error) {
	path := "{{ .Path }}"

	{{- if .Parameters }}
	// Replace path parameters and prepare query parameters
	queryParams := make(map[string]string)
	{{- range .Parameters }}
	{{- if eq .In "path" }}
	path = strings.ReplaceAll(path, "{{ printf "{%s}" .JSONName }}", fmt.Sprintf("%v", params.{{ .Name }}))
	{{- else if eq .In "query" }}
	if lo.IsNotEmpty(params.{{ .Name }}) {
		queryParams["{{ .JSONName }}"] = fmt.Sprintf("%v", params.{{ .Name }})
	}
	{{- end }}
	{{- end }}
	{{- end }}

	// Create request
	req := c.client.R()

	{{- if .Parameters }}
	// Set query parameters
	req.SetQueryParams(queryParams)

	{{- range .Parameters }}
	{{- if eq .In "header" }}
	// Set header parameters
	if lo.IsNotEmpty(params.{{ .Name }}) {
		req.SetHeader("{{ .JSONName }}", fmt.Sprintf("%v", params.{{ .Name }}))
	}
	{{- end }}
	{{- end }}
	{{- end }}

	{{- if .Body }}
	// Set request body
	req.SetBody(body)
	{{- end }}

	{{- if .Response }}
	// Set response object
	var result {{ .Response.Name }}
	req.SetResult(&result)
	{{- end }}

	// Send request
	resp, err := req.{{ .Method }}(path)
	if err != nil {
		return {{ if .Response }}nil, {{ end }}fmt.Errorf("error sending request: %w", err)
	}

	// Check for successful status code
	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return {{ if .Response }}nil, {{ end }}fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode(), resp.String())
	}

	{{- if .Response }}
	return &result, nil
	{{- else }}
	return nil
	{{- end }}
}
{{ end }}
```

so that you can generate the client code with the following command:
```bash
openapi-codegen -i openapi.yaml -o . -p api -t client.tmpl
```
