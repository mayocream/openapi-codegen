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
	// Parameter
	In string
}

// FileData represents all structs for the generated Go file
type FileData struct {
	PackageName string
	Schemas     []*Schema
	Requests    []*Request
}

// Request represents a single request to be generated
type Request struct {
	Name       string
	Path       string
	Method     string
	Parameters []*Schema
	Body       *Schema
	Response   *Schema
}
