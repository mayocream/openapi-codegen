package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
)

// generateSchema generates Go types from OpenAPI schema
func generateSchema(schemaRef *openapi3.SchemaRef, schemaName string) (*Schema, []*Schema, error) {
	if schemaRef == nil || schemaRef.Value == nil {
		return nil, nil, fmt.Errorf("schema is nil")
	}

	schema := &Schema{
		Name:        lo.PascalCase(schemaName),
		Description: schemaRef.Value.Description,
	}

	if schemaRef.Ref != "" {
		schema.Type = extractTypeNameFromRef(schemaRef.Ref)
		return schema, nil, nil
	}

	var additionalSchemas []*Schema

	switch {
	case schemaRef.Value.Type.Is(openapi3.TypeObject) && schemaRef.Value.Properties != nil:
		schema.Type = "struct"
		schema.Properties = lo.MapToSlice(schemaRef.Value.Properties, func(propName string, propSchema *openapi3.SchemaRef) *Schema {
			fieldName := lo.PascalCase(propName)
			fieldSchema, nestedSchemas, _ := generateSchema(propSchema, fieldName)
			additionalSchemas = append(additionalSchemas, nestedSchemas...)
			fieldSchema.Name = fieldName
			fieldSchema.JSONName = propName
			fieldSchema.OmitEmpty = !lo.Contains(schemaRef.Value.Required, propName)
			return fieldSchema
		})
		// Sort properties by name
		slices.SortStableFunc(schema.Properties, func(a, b *Schema) int {
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

// generateResponseSchema generates Go types from OpenAPI response
func generateResponseSchema(response *openapi3.ResponseRef, schemaName string) (*Schema, []*Schema, error) {
	if response == nil || response.Value == nil {
		return nil, nil, fmt.Errorf("response is nil")
	}

	for mediaType, mediaTypeValue := range response.Value.Content {
		if mediaType == "application/json" && mediaTypeValue.Schema != nil {
			return generateSchema(mediaTypeValue.Schema, schemaName)
		}
	}

	return &Schema{
		Name:        lo.PascalCase(schemaName),
		Type:        "any",
		Description: *response.Value.Description,
	}, nil, nil
}

// generateRequests generates Go types from OpenAPI request parameter
func generateRequests(pathItem *openapi3.PathItem, path string) ([]*Request, error) {
	var requests []*Request
	var pathSchemas []*Schema

	findParameter := func(name string) *openapi3.ParameterRef {
		return spec.Components.Parameters[name]
	}

	// Create path parameters
	if pathItem.Parameters != nil {
		for _, param := range pathItem.Parameters {
			// Skip parameters that are not in path
			if param.Ref == "" {
				continue
			}
			if p := findParameter(param.Ref); p != nil {
				schema, _, _ := generateSchema(p.Value.Schema, p.Value.Name)
                schema.In = p.Value.In
				pathSchemas = append(pathSchemas, schema)
			}
		}
	}

	// Create requests for each HTTP method
	for method, operation := range pathItem.Operations() {
		request := &Request{
			Name: lo.PascalCase(operation.OperationID),
			Path: path,
            Method: lo.Capitalize(method),
            Parameters: make([]*Schema, 0),
		}

        // Add path parameters
        request.Parameters = append(request.Parameters, pathSchemas...)

        // other parameters
        if operation.Parameters != nil {
            for _, param := range operation.Parameters {
                if param.Ref == "" {
                    continue
                }
                if p := findParameter(param.Ref); p != nil {
                    schema, _, _ := generateSchema(p.Value.Schema, p.Value.Name)
                    schema.In = p.Value.In
                    request.Parameters = append(request.Parameters, schema)
                }
            }
        }

        requests = append(requests, request)
	}

	return requests, nil
}

func extractTypeNameFromRef(ref string) string {
	return lo.PascalCase(lo.LastOrEmpty(strings.Split(ref, "/")))
}

// Generate Go code for all components
func generateComponents(spec *openapi3.T, packageName string) (string, error) {
	fileData := FileData{
		PackageName: packageName,
		Schemas:     make([]*Schema, 0),
	}

	// Components.Schemas
	{
		// ordering keys
		keys := getYAMLNodeKeys("components.schemas")
		if keys == nil {
			return "", fmt.Errorf("failed to get components.schemas keys")
		}

		for _, key := range keys {
			ref := spec.Components.Schemas[key]
			schema, additionalSchemas, err := generateSchema(ref, key)
			if err != nil {
				return "", fmt.Errorf("failed to generate schema for %s: %w", key, err)
			}
			fileData.Schemas = append(fileData.Schemas, schema)
			fileData.Schemas = append(fileData.Schemas, additionalSchemas...)
		}
	}
	// Components.Responses
	{
		keys := getYAMLNodeKeys("components.responses")
		if keys == nil {
			return "", fmt.Errorf("failed to get components.responses keys")
		}

		for _, key := range keys {
			ref := spec.Components.Responses[key]
			schema, additionalSchemas, err := generateResponseSchema(ref, key)
			if err != nil {
				return "", fmt.Errorf("failed to generate schema for %s: %w", key, err)
			}
			fileData.Schemas = append(fileData.Schemas, schema)
			fileData.Schemas = append(fileData.Schemas, additionalSchemas...)
		}
	}

	// Remove duplicates
	fileData.Schemas = lo.UniqBy(fileData.Schemas, func(s *Schema) string {
		return s.Name
	})

	return applySchemaTemplate(fileData)
}

func generateClient(spec *openapi3.T, packageName string) (string, error) {
    fileData := FileData{
		PackageName: packageName,
		Requests:    make([]*Request, 0),
	}

    keys := getYAMLNodeKeys("paths")
    if keys == nil {
        return "", fmt.Errorf("failed to get paths keys")
    }

    for _, key := range keys {
        pathItem := spec.Paths.Value(key)
        requests, err := generateRequests(pathItem, key)
        if err != nil {
            return "", fmt.Errorf("failed to generate requests for %s: %w", key, err)
        }
        fileData.Requests = append(fileData.Requests, requests...)
    }

    return applyClientTemplate(fileData)
}
