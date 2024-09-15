package main

import (
	"testing"
)

func Test_generateComponents(t *testing.T) {
	_, _, err := parseOpenAPISpec("testdata/openapi.yaml")
	if err != nil {
		t.Errorf("parseOpenAPISpec() error = %v", err)
		return
	}
	_, err = generateComponents(spec, "testdata")
	if err != nil {
		t.Errorf("generateComponents() error = %v", err)
		return
	}
}
