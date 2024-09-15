package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

var (
	node *yaml.Node
	spec *openapi3.T
)

// Load and parse OpenAPI 3.0 spec
func parseOpenAPISpec(filePath string) (*yaml.Node, *openapi3.T, error) {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read OpenAPI spec file: %w", err)
	}

	node = &yaml.Node{}
	if err := yaml.Unmarshal(raw, node); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	loader := openapi3.NewLoader()
	spec, err = loader.LoadFromData(raw)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}
	return node, spec, nil
}

// getYAMLNodeKeys returns the keys of a YAML node
// Example: getYAMLNodeKeys("components.schemas") -> ["UserExists", "Response", "Error", ...]
func getYAMLNodeKeys(nodeKey string) []string {
	v := node.Content[0]
	keys := strings.Split(nodeKey, ".")

	for _, key := range keys {
		if v.Kind != yaml.MappingNode {
			return nil
		}
		v = lo.FindOrElse(lo.Chunk(v.Content, 2), nil, func(pair []*yaml.Node) bool {
			return pair[0].Value == key
		})[1]
		if v == nil {
			return nil
		}
	}

	if v.Kind != yaml.MappingNode {
		return nil
	}

	return lo.Map(lo.Chunk(v.Content, 2), func(pair []*yaml.Node, _ int) string {
		return pair[0].Value
	})
}
