package runner

import (
	"errors"

	"go.yaml.in/yaml/v3"
)

type yamlScalarExpectation struct {
	path []string
}

func decodeYAMLScalars(node *yaml.Node) ([]yamlScalarExpectation, error) {
	if err := validateYAMLNodes(node); err != nil {
		return nil, err
	}
	if node.Kind != yaml.MappingNode || len(node.Content) != 2 || node.Content[0].Value != "yaml-scalars" {
		return nil, errors.New("yaml-scalars must be the only file assertion")
	}
	items := node.Content[1]
	if items.Kind != yaml.SequenceNode || len(items.Content) == 0 {
		return nil, errors.New("yaml-scalars must be a nonempty sequence")
	}
	result := make([]yamlScalarExpectation, 0, len(items.Content))
	for _, item := range items.Content {
		parsed, err := decodeYAMLScalar(item)
		if err != nil {
			return nil, err
		}
		result = append(result, parsed)
	}
	return result, nil
}

func decodeYAMLScalar(node *yaml.Node) (yamlScalarExpectation, error) {
	if node.Kind != yaml.MappingNode || len(node.Content) != 4 {
		return yamlScalarExpectation{}, errors.New("a yaml-scalars predicate requires path and utc-timestamp")
	}
	var path, timestamp *yaml.Node
	for index := 0; index < len(node.Content); index += 2 {
		switch node.Content[index].Value {
		case "path":
			path = node.Content[index+1]
		case "utc-timestamp":
			timestamp = node.Content[index+1]
		default:
			return yamlScalarExpectation{}, errors.New("unknown yaml-scalars predicate field")
		}
	}
	if path == nil || timestamp == nil {
		return yamlScalarExpectation{}, errors.New("a yaml-scalars predicate requires path and utc-timestamp")
	}
	var enabled bool
	if timestamp.Kind != yaml.ScalarNode || timestamp.ShortTag() != "!!bool" || timestamp.Decode(&enabled) != nil || !enabled {
		return yamlScalarExpectation{}, errors.New("utc-timestamp must be the YAML boolean true")
	}
	if path.Kind != yaml.SequenceNode || len(path.Content) == 0 {
		return yamlScalarExpectation{}, errors.New("path must be a nonempty sequence")
	}
	result := yamlScalarExpectation{path: make([]string, 0, len(path.Content))}
	for _, item := range path.Content {
		if item.Kind != yaml.ScalarNode || item.ShortTag() != "!!str" || item.Value == "" {
			return yamlScalarExpectation{}, errors.New("path items must be nonempty YAML strings")
		}
		result.path = append(result.path, item.Value)
	}
	return result, nil
}

func validateYAMLNodes(root *yaml.Node) error {
	pending := []*yaml.Node{root}
	for len(pending) > 0 {
		node := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if node.Anchor != "" || node.Kind == yaml.AliasNode {
			return errors.New("YAML anchors and aliases are unsupported")
		}
		if node.Kind == yaml.MappingNode {
			keys := make(map[string]bool, len(node.Content)/2)
			for index := 0; index+1 < len(node.Content); index += 2 {
				key := node.Content[index]
				if key.Kind != yaml.ScalarNode || key.ShortTag() != "!!str" {
					return errors.New("YAML mapping keys must be strings")
				}
				if keys[key.Value] {
					return errors.New("duplicated YAML mapping key")
				}
				keys[key.Value] = true
			}
		}
		pending = append(pending, node.Content...)
	}
	return nil
}
