# openapi-codegen

Implementation of OpenAPI code generation in Go.

Supports [OpenAPI 3.0.3](https://spec.openapis.org/oas/v3.0.3.html).

## Features

- [x] Type-safe structs generation, based on [kin-openapi](https://github.com/getkin/kin-openapi) package
- [ ] HTTP Client generation, based on [resty](https://github.com/go-resty/resty) package

## Usage

```bash
go install github.com/mayocream/openapi-codegen@latest

openapi-codegen -i <input-file> -o <output-dir> -p <package-name>
```
