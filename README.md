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

openapi-codegen -i openapi.yaml -o . -p api -t client.tmpl
```
