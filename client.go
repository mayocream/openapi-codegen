package main

import (
	"fmt"
	"go/format"
	"log"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
)

const clientTemplate = `package vrchatgo

import (
	"fmt"
	"net/http"
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client *resty.Client
}

func NewClient() *Client {
	return &Client{
		client: resty.New().SetBaseURL("https://vrchatapi.com/api/1"),
	}
}

{{range .}}
{{if .HasBody}}
func (c *Client) {{.MethodName}}({{.Params}}) ({{.ResponseType}}, error) {
	resp, err := c.client.R().
		SetBody({{.RequestBody}}).
		{{.Method}}("{{.Path}}")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("API error: %v", resp.Status())
	}
	var result {{.ResponseType}}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
{{else}}
func (c *Client) {{.MethodName}}({{.Params}}) ({{.ResponseType}}, error) {
	resp, err := c.client.R().
		{{.Method}}("{{.Path}}")
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("API error: %v", resp.Status())
	}
	var result {{.ResponseType}}
	err = json.Unmarshal(resp.Body(), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
{{end}}
{{end}}
`

type MethodInfo struct {
	MethodName   string
	Params       string
	Method       string
	Path         string
	ResponseType string
	HasBody      bool
	RequestBody  string
}

func GenerateClient(spec *openapi3.T) (string, error) {
	var methods []MethodInfo

	for _, path := range spec.Paths.InMatchingOrder() {
		pathItem := spec.Paths.Value(path)
		for method, operation := range pathItem.Operations() {
			methodInfo := MethodInfo{
				MethodName:   lo.PascalCase(operation.OperationID),
				Method:       lo.Capitalize(method),
				Path:         path,
				ResponseType: getResponseType(operation.Responses),
				HasBody:      hasRequestBody(operation.RequestBody),
			}

			methodInfo.Params = generateParams(operation.Parameters, methodInfo.HasBody)
			if methodInfo.HasBody {
				methodInfo.RequestBody = "requestBody"
			}

			methods = append(methods, methodInfo)
		}
	}

	tmpl, err := template.New("client").Parse(clientTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	err = tmpl.Execute(&buf, methods)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil

	formattedCode, err := format.Source([]byte(buf.String()))
	if err != nil {
		return "", fmt.Errorf("failed to format code: %w", err)
	}

	return string(formattedCode), nil
}

func getResponseType(responses *openapi3.Responses) string {
	for _, response := range responses.Map() {
		if response.Value != nil && response.Value.Content != nil {
			for mediaType, mediaTypeValue := range response.Value.Content {
				if mediaType == "application/json" && mediaTypeValue.Schema != nil {
					return getSchemaType(mediaTypeValue.Schema.Value)
				}
			}
		}
	}
	return "interface{}"
}

func getSchemaType(schema *openapi3.Schema) string {
	if schema.Type == nil {
		return "interface{}"
	}

	switch (*schema.Type)[0] {
	case "object":
		return "map[string]interface{}"
	case "array":
		if schema.Items != nil {
			return "[]" + getSchemaType(schema.Items.Value)
		}
		return "[]interface{}"
	case "string":
		return "string"
	case "integer":
		return "int"
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "interface{}"
	}
}

func hasRequestBody(requestBody *openapi3.RequestBodyRef) bool {
	return requestBody != nil && requestBody.Value != nil
}

func generateParams(parameters openapi3.Parameters, hasBody bool) string {
	var params []string
	for _, param := range parameters {
		if param.Value != nil {
			paramName := param.Value.Name
			paramType := getSchemaType(param.Value.Schema.Value)
			params = append(params, fmt.Sprintf("%s %s", paramName, paramType))
		}
	}
	if hasBody {
		params = append(params, "requestBody interface{}")
	}
	return strings.Join(params, ", ")
}

func GenerateVRChatClient(specFile string) (string, error) {
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(specFile)
	if err != nil {
		return "", fmt.Errorf("failed to load spec: %w", err)
	}

	log.Printf("loaded spec: %v", spec.Info.Title)

	return GenerateClient(spec)
}
