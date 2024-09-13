package openapicodegen

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
)

// Struct template for Go types with description and 4 spaces indentation
const fileTemplate = `package {{ .PackageName }}

{{ range .Structs }}
type {{ .StructName }} struct {
{{- range .Fields }}
    {{ .Description }}{{ .FieldName }} {{ .FieldType | indentType }} ` + "`json:\"{{ .JSONName }}\"`" + `
{{- end }}
}
{{ end }}
`

// Template function to handle indentation for inline types
func indentType(typeStr string) string {
	lines := strings.Split(typeStr, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = "    " + lines[i]
	}
	return strings.Join(lines, "\n")
}

// Load and parse OpenAPI 3.0 spec.
func parseOpenAPISpec(filePath string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	return spec, nil
}

// Field represents a single struct field
type Field struct {
	FieldName   string
	FieldType   string
	JSONName    string
	Description string
}

// StructData represents a Go struct to be generated
type StructData struct {
	StructName string
	Fields     []Field
}

// FileData represents all structs for the generated Go file
type FileData struct {
	PackageName string
	Structs     []StructData
}

// Generate Go type for a single schema
func generateSchema(schema *openapi3.Schema, schemaName string, spec *openapi3.T) (StructData, error) {
	structData := StructData{
		StructName: lo.PascalCase(schemaName),
		Fields:     []Field{},
	}

	if schema == nil {
		return structData, fmt.Errorf("schema is nil")
	}

	// If the schema is of type "object", generate a struct
	if schema.Type.Is("object") {
		lo.ForEach(lo.Keys(schema.Properties), func(propName string, _ int) {
			propSchema := schema.Properties[propName]
			fieldName := lo.PascalCase(propName) // PascalCase for property name
			fieldType, err := resolveGoType(propSchema, spec)
			if err == nil {
				description := ""
				if propSchema.Value != nil && strings.TrimSpace(propSchema.Value.Description) != "" {
					description = fmt.Sprintf("// %s %s\n    ", fieldName, strings.ReplaceAll(propSchema.Value.Description, "\n", "\n    // "))
				}
				structData.Fields = append(structData.Fields, Field{
					FieldName:   fieldName,
					FieldType:   fieldType,
					JSONName:    propName,
					Description: description,
				})
			}
		})
	}

	return structData, nil
}

// Resolve Go type for a given schema reference or inline schema
func resolveGoType(schemaRef *openapi3.SchemaRef, spec *openapi3.T) (string, error) {
	if schemaRef.Ref != "" {
		refName := extractTypeNameFromRef(schemaRef.Ref)
		return refName, nil
	}

	if schemaRef.Value.Type != nil {
		switch (*schemaRef.Value.Type)[0] {
		case "string": // string
			return "string", nil
		case "integer": // integer
			return "int", nil
		case "number": // number
			return "float64", nil
		case "boolean": // boolean
			return "bool", nil
		case "array": // array
			itemType, err := resolveGoType(schemaRef.Value.Items, spec)
			if err != nil {
				return "", err
			}
			return "[]" + itemType, nil
		case "object": // object
			inlineStruct, err := generateSchema(schemaRef.Value, "Inline", spec)
			if err != nil {
				return "", err
			}
			// If the inline struct is empty, we just return an empty struct type.
			if len(inlineStruct.Fields) == 0 {
				return "struct {}", nil
			}
			// Correct formatting of inline struct with fields.
			return fmt.Sprintf("struct {\n%s\n}", generateInlineStruct(inlineStruct.Fields)), nil
		}
	}

	return "interface{}", nil
}

// Extract the Go type name from an OpenAPI $ref (reference) string
func extractTypeNameFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	return parts[len(parts)-1] // Return the last part which is the type name
}

// Helper function to generate the inline struct fields as Go code
func generateInlineStruct(fields []Field) string {
	var result strings.Builder
	for _, field := range fields {
		result.WriteString(fmt.Sprintf("    %s %s `json:\"%s\"`\n", field.FieldName, field.FieldType, field.JSONName))
	}
	return result.String()
}

// Apply Go templates to generate code
func applyTemplate(tmpl string, data any) (string, error) {
	var buf bytes.Buffer
	t, err := template.New("goFile").Funcs(template.FuncMap{
		"indentType": indentType, // Template function to apply indentation
	}).Parse(tmpl)
	if err != nil {
		return "", err
	}
	err = t.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Generate Go code for all components
func generateComponents(spec *openapi3.T, packageName string) (string, error) {
	fileData := FileData{
		PackageName: packageName,
		Structs:     []StructData{},
	}

	lo.ForEach(lo.Keys(spec.Components.Schemas), func(name string, _ int) {
		schema := spec.Components.Schemas[name]
		structData, err := generateSchema(schema.Value, name, spec)
		if err == nil {
			fileData.Structs = append(fileData.Structs, structData)
		}
	})

	return applyTemplate(fileTemplate, fileData)
}

// Generate Go code from the OpenAPI spec
func generate(spec *openapi3.T, packageName, outputFilePath string) error {
	code, err := generateComponents(spec, packageName)
	if err != nil {
		return fmt.Errorf("error generating components: %w", err)
	}

	return os.WriteFile(outputFilePath, []byte(code), 0644)
}

func main() {
	// Parse the OpenAPI spec
	spec, err := parseOpenAPISpec("openapi.yaml")
	if err != nil {
		fmt.Printf("Error parsing OpenAPI spec: %v\n", err)
		return
	}

	// Generate Go code from the spec
	err = generate(spec, "my_package", "generated.go")
	if err != nil {
		fmt.Printf("Error generating code: %v\n", err)
	}
}
