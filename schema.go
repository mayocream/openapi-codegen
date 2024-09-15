package main

import (
	"github.com/getkin/kin-openapi/openapi3"
	"gopkg.in/yaml.v3"
)

// Doc represents a parsed OpenAPI 3.0 spec
type Doc struct {
	Raw  *yaml.Node
	Spec *openapi3.T
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
	Properties []*Schema
	// Enum
	EnumValues []any
}

// FileData represents all structs for the generated Go file
type FileData struct {
	PackageName string
	Schemas     []*Schema
	Methods     []MethodInfo
}

type ParameterInfo struct {
	Name     string
	Type     string
	Required bool
	In       string
}

type MethodInfo struct {
	MethodName      string
	ParamsStruct    string
	Method          string
	Path            string
	ResponseType    string
	HasBody         bool
	RequestBody     string
	Parameters      []ParameterInfo
	QueryParams     []ParameterInfo
	PathParams      []ParameterInfo
	HeaderParams    []ParameterInfo
	HasQueryParams  bool
	HasPathParams   bool
	HasHeaderParams bool
}

type ClientFileData struct {
	PackageName string
	Methods     []MethodInfo
}
