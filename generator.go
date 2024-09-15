package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
)

// generateSchema generates Go types from OpenAPI schema
func generateSchema(schemaRef *openapi3.SchemaRef, schemaName string) (Schema, []Schema, error) {
    if schemaRef == nil || schemaRef.Value == nil {
        return Schema{}, nil, fmt.Errorf("schema is nil")
    }

    schema := Schema{
        Name:        lo.PascalCase(schemaName),
        Description: schemaRef.Value.Description,
    }

    if schemaRef.Ref != "" {
        schema.Type = extractTypeNameFromRef(schemaRef.Ref)
        return schema, nil, nil
    }

    var additionalSchemas []Schema

    switch {
    case schemaRef.Value.Type.Is(openapi3.TypeObject) && schemaRef.Value.Properties != nil:
        schema.Type = "struct"
        schema.Properties = lo.MapToSlice(schemaRef.Value.Properties, func(propName string, propSchema *openapi3.SchemaRef) Schema {
            fieldName := lo.PascalCase(propName)
            fieldSchema, nestedSchemas, _ := generateSchema(propSchema, fieldName)
            additionalSchemas = append(additionalSchemas, nestedSchemas...)
            fieldSchema.Name = fieldName
            fieldSchema.JSONName = propName
            fieldSchema.OmitEmpty = !lo.Contains(schemaRef.Value.Required, propName)
            return fieldSchema
        })
		// Sort properties by name
		slices.SortStableFunc(schema.Properties, func(a, b Schema) int {
			return strings.Compare(a.Name, b.Name)
		})

    case schemaRef.Value.Enum != nil:
        schema.Type = "enum"
        schema.EnumValues = schemaRef.Value.Enum

    case schemaRef.Value.Type.Is(openapi3.TypeArray):
        itemSchema, nestedSchemas, _ := generateSchema(schemaRef.Value.Items, schemaName+"Item")
        schema.Type = "[]" + itemSchema.Type
        additionalSchemas = append(additionalSchemas, nestedSchemas...)

    case schemaRef.Value.Type.Is(openapi3.TypeString):
        if schemaRef.Value.Format == "date-time" {
            schema.Type = "time.Time"
        } else {
            schema.Type = "string"
        }

    case schemaRef.Value.Type.Is(openapi3.TypeInteger):
        schema.Type = "int64"

    case schemaRef.Value.Type.Is(openapi3.TypeNumber):
        schema.Type = "float64"

    case schemaRef.Value.Type.Is(openapi3.TypeBoolean):
        schema.Type = "bool"

    default:
        schema.Type = "any"
    }

    return schema, additionalSchemas, nil
}

func extractTypeNameFromRef(ref string) string {
	return lo.PascalCase(lo.LastOrEmpty(strings.Split(ref, "/")))
}

// Generate Go code for all components
func generateComponents(spec *openapi3.T, packageName string) (string, error) {
	fileData := FileData{
		PackageName: packageName,
		Schemas:     []Schema{},
	}

	// ordering keys
	keys := getYAMLNodeKeys("components.schemas")
	if keys == nil {
		return "", fmt.Errorf("failed to get components.schemas keys")
	}

	for _, key := range keys {
		schemaRef := spec.Components.Schemas[key]
		schema, additionalSchemas, err := generateSchema(schemaRef, key)
		if err != nil {
			return "", fmt.Errorf("failed to generate schema for %s: %w", key, err)
		}
		fileData.Schemas = append(fileData.Schemas, schema)
		fileData.Schemas = append(fileData.Schemas, additionalSchemas...)
	}

	return applySchemaTemplate(fileData)
}

