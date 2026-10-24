package runner

import (
	"bytes"
	"strings"

	"go.yaml.in/yaml/v3"
)

func extensionInDocument(data []byte) bool {
	var document yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&document) != nil || document.Kind != yaml.DocumentNode || len(document.Content) != 1 {
		return false
	}
	steps := documentField(document.Content[0], "steps")
	steps = documentValue(steps)
	if steps == nil || steps.Kind != yaml.SequenceNode {
		return false
	}
	for _, step := range steps.Content {
		if documentField(step, "capture") != nil || documentInterpolation(documentField(step, "stdin")) {
			return true
		}
		if documentSequenceInterpolation(documentField(step, "run")) || documentMapInterpolation(documentField(step, "files")) {
			return true
		}
		expect := documentField(step, "expect")
		if documentSequenceInterpolation(documentField(expect, "unchanged")) {
			return true
		}
		files := documentValue(documentField(expect, "files"))
		if documentMapInterpolation(files) {
			return true
		}
		if files == nil || files.Kind != yaml.MappingNode {
			continue
		}
		for index := 1; index < len(files.Content); index += 2 {
			predicates := documentValue(documentField(files.Content[index], "yaml-scalars"))
			if predicates == nil || predicates.Kind != yaml.SequenceNode {
				continue
			}
			for _, predicate := range predicates.Content {
				if documentField(predicate, "equals") != nil || documentField(predicate, "absent") != nil {
					return true
				}
			}
		}
	}
	return false
}

func documentValue(node *yaml.Node) *yaml.Node {
	seen := map[*yaml.Node]bool{}
	for node != nil && node.Kind == yaml.AliasNode {
		if seen[node] {
			return nil
		}
		seen[node] = true
		node = node.Alias
	}
	return node
}

func documentField(node *yaml.Node, name string) *yaml.Node {
	node = documentValue(node)
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for index := 0; index+1 < len(node.Content); index += 2 {
		key := node.Content[index]
		if key.Kind == yaml.ScalarNode && key.ShortTag() == "!!str" && key.Value == name {
			return node.Content[index+1]
		}
	}
	return nil
}

func documentInterpolation(node *yaml.Node) bool {
	node = documentValue(node)
	return node != nil && node.Kind == yaml.ScalarNode && strings.Contains(node.Value, "${")
}

func documentSequenceInterpolation(node *yaml.Node) bool {
	node = documentValue(node)
	if node != nil && node.Kind == yaml.SequenceNode {
		for _, item := range node.Content {
			if documentInterpolation(item) {
				return true
			}
		}
	}
	return false
}

func documentMapInterpolation(node *yaml.Node) bool {
	node = documentValue(node)
	if node != nil && node.Kind == yaml.MappingNode {
		for index := 0; index+1 < len(node.Content); index += 2 {
			if documentInterpolation(node.Content[index]) {
				return true
			}
		}
	}
	return false
}
