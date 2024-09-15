package main

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
)


func generateClient(spec *openapi3.T, packageName string) (string, error) {
	var methods []MethodInfo

	for _, path := range spec.Paths.InMatchingOrder() {
		pathItem := spec.Paths.Value(path)
		for method, operation := range pathItem.Operations() {
			methodInfo := MethodInfo{
				MethodName:   lo.PascalCase(operation.OperationID),
				Method:       strings.ToUpper(method),
				Path:         path,
				ResponseType: getResponseType(operation.Responses),
				HasBody:      hasRequestBody(operation.RequestBody),
			}

			methodInfo.Parameters = generateParameters(operation.Parameters)
			methodInfo.QueryParams = filterParameters(methodInfo.Parameters, "query")
			methodInfo.PathParams = filterParameters(methodInfo.Parameters, "path")
			methodInfo.HeaderParams = filterParameters(methodInfo.Parameters, "header")
			methodInfo.HasQueryParams = len(methodInfo.QueryParams) > 0
			methodInfo.HasPathParams = len(methodInfo.PathParams) > 0
			methodInfo.HasHeaderParams = len(methodInfo.HeaderParams) > 0

			methodInfo.ParamsStruct = generateParamsStruct(methodInfo)

			if methodInfo.HasBody {
				methodInfo.RequestBody = getRequestBodyType(operation.RequestBody)
			}

			methods = append(methods, methodInfo)
		}
	}

	clientData := ClientFileData{
		PackageName: packageName,
		Methods:     methods,
	}

	return applyClientTemplate(clientData)
}

func generateParameters(parameters openapi3.Parameters) []ParameterInfo {
	var params []ParameterInfo
	for _, param := range parameters {
		if param.Value != nil {
			params = append(params, ParameterInfo{
				Name:     param.Value.Name,
				Type:     getSchemaType(param.Value.Schema),
				Required: param.Value.Required,
				In:       param.Value.In,
			})
		}
	}
	return params
}

func filterParameters(params []ParameterInfo, in string) []ParameterInfo {
	return lo.Filter(params, func(p ParameterInfo, _ int) bool {
		return p.In == in
	})
}

func generateParamsStruct(methodInfo MethodInfo) string {
	var fields []string
	for _, param := range methodInfo.Parameters {
		fields = append(fields, fmt.Sprintf("%s %s `json:\"%s\"`", lo.PascalCase(param.Name), param.Type, param.Name))
	}
	if methodInfo.HasBody {
		fields = append(fields, fmt.Sprintf("RequestBody %s `json:\"requestBody\"`", methodInfo.RequestBody))
	}
	return fmt.Sprintf("type %sParams struct {\n\t%s\n}", methodInfo.MethodName, strings.Join(fields, "\n\t"))
}

func getResponseType(responses *openapi3.Responses) string {
	for _, response := range responses.Map() {
		if response.Value != nil && response.Value.Content != nil {
			for mediaType, mediaTypeValue := range response.Value.Content {
				if mediaType == "application/json" && mediaTypeValue.Schema != nil {
					return getSchemaType(mediaTypeValue.Schema)
				}
			}
		}
	}
	return "interface{}"
}

func getSchemaType(schemaRef *openapi3.SchemaRef) string {
	if schemaRef.Ref != "" {
		return extractTypeNameFromRef(schemaRef.Ref)
	}

	schema := schemaRef.Value
	if schema.Type == nil {
		return "interface{}"
	}

	switch (*schema.Type)[0] {
	case "object":
		return "map[string]interface{}"
	case "array":
		if schema.Items != nil {
			return "[]" + getSchemaType(schema.Items)
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

func getRequestBodyType(requestBody *openapi3.RequestBodyRef) string {
	if requestBody != nil && requestBody.Value != nil {
		for mediaType, mediaTypeValue := range requestBody.Value.Content {
			if mediaType == "application/json" && mediaTypeValue.Schema != nil {
				return getSchemaType(mediaTypeValue.Schema)
			}
		}
	}
	return "interface{}"
}
