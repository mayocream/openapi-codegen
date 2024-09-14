package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
)

// Load and parse OpenAPI 3.0 spec
func parseOpenAPISpec(filePath string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	spec, err := loader.LoadFromFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	return spec, nil
}

// Schema represents a single schema to be generated
type Schema struct {
	Name        string
	Type        string
	JSONName    string
	Description string
	OmitEmpty   bool
	Example     string
	// Object
	Properties []Schema
}

// FileData represents all structs for the generated Go file
type FileData struct {
	PackageName string
	Schemas     []Schema
}

// Generate Go type for a single schema
func generateSchema(schemaDef *openapi3.Schema, schemaName string, spec *openapi3.T) (Schema, []Schema, error) {
	schema := Schema{
		Name: lo.PascalCase(schemaName),
	}

	var additionalSchemas []Schema

	if schemaDef == nil {
		return schema, additionalSchemas, fmt.Errorf("schema is nil")
	}

	schema.Description = schemaDef.Description

	if schemaDef.Type.Is("object") {
		schema.Type = "struct"
		for propName, propSchema := range schemaDef.Properties {
			fieldName := lo.PascalCase(propName)
			fieldType, nestedSchemas, err := resolveGoType(propSchema, fieldName, spec)
			if err == nil {
				schema.Properties = append(schema.Properties, Schema{
					Name:        fieldName,
					Type:        fieldType,
					JSONName:    propName,
					Description: propSchema.Value.Description,
					OmitEmpty:   !lo.Contains(schemaDef.Required, propName),
				})
				additionalSchemas = append(additionalSchemas, nestedSchemas...)
			}
		}
	} else {
		var nestedSchemas []Schema
		schema.Type, nestedSchemas, _ = resolveGoType(&openapi3.SchemaRef{Value: schemaDef}, schemaName, spec)
		additionalSchemas = append(additionalSchemas, nestedSchemas...)
	}

	return schema, additionalSchemas, nil
}

// Resolve Go type for a given schema reference or inline schema
func resolveGoType(schemaRef *openapi3.SchemaRef, parentName string, spec *openapi3.T) (string, []Schema, error) {
	if schemaRef.Ref != "" {
		return extractTypeNameFromRef(schemaRef.Ref), nil, nil
	}

	var additionalSchemas []Schema

	if schemaRef.Value.Type != nil {
		switch (*schemaRef.Value.Type)[0] {
		case "string":
			if schemaRef.Value.Format == "date-time" {
				return "time.Time", nil, nil
			}
			// TODO: enum type
			return "string", nil, nil
		case "integer":
			return "int", nil, nil
		case "number":
			return "float64", nil, nil
		case "boolean":
			return "bool", nil, nil
		case "array":
			itemType, nestedSchemas, err := resolveGoType(schemaRef.Value.Items, parentName+"Item", spec)
			if err != nil {
				return "", nil, err
			}
			additionalSchemas = append(additionalSchemas, nestedSchemas...)
			return "[]" + itemType, additionalSchemas, nil
		case "object":
			nestedSchemaName := parentName + "Nested"
			nestedSchema, nestedSchemas, err := generateSchema(schemaRef.Value, nestedSchemaName, spec)
			if err != nil {
				return "", nil, err
			}
			additionalSchemas = append(additionalSchemas, nestedSchema)
			additionalSchemas = append(additionalSchemas, nestedSchemas...)
			return nestedSchemaName, additionalSchemas, nil
		}
	}

	return "any", nil, nil
}

// Extract the Go type name from an OpenAPI $ref (reference) string
func extractTypeNameFromRef(ref string) string {
	parts := strings.Split(ref, "/")
	return lo.PascalCase(parts[len(parts)-1])
}

// Generate Go code for all components
func generateComponents(spec *openapi3.T, packageName string) (string, error) {
	fileData := FileData{
		PackageName: packageName,
		Schemas:     []Schema{},
	}

	for name, schemaDef := range spec.Components.Schemas {
		schema, additionalSchemas, err := generateSchema(schemaDef.Value, name, spec)
		if err == nil {
			fileData.Schemas = append(fileData.Schemas, schema)
			fileData.Schemas = append(fileData.Schemas, additionalSchemas...)
		}
	}

	return applySchemaTemplate(fileData)
}

// Generate Go code from the OpenAPI spec
func generate(spec *openapi3.T, packageName, outputFilePath string) error {
	code, err := generateComponents(spec, packageName)
	if err != nil {
		return fmt.Errorf("error generating components: %w", err)
	}

	return os.WriteFile(outputFilePath, []byte(code), 0644)
}
