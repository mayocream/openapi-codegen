package openapicodegen

import (
	"testing"

	"github.com/k0kubun/pp/v3"
)

func Test_parseOpenAPISpec(t *testing.T) {
	spec, err := parseOpenAPISpec("testdata/openapi.yaml")
	if err != nil {
		t.Fatalf("parseOpenAPISpec() error = %v", err)
	}
	t.Logf("spec: %v", pp.Sprint(spec))
}

func Test_generate(t *testing.T) {
	spec, err := parseOpenAPISpec("testdata/openapi.yaml")
	if err != nil {
		t.Fatalf("parseOpenAPISpec() error = %v", err)
	}
	if err := generate(spec, "testdata", "testdata/openapi.gen.go"); err != nil {
		t.Errorf("generate() error = %v", err)
	}
}
