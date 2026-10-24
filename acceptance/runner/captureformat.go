package runner

import (
	"errors"
	"regexp"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Capture struct {
	UniqueDirectoryEntry *string
	YAMLScalar           *YAMLScalarCapture
}

type YAMLScalarCapture struct {
	File string
	Path []string
}

var captureName = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,63}$`)

func (capture *Capture) UnmarshalYAML(node *yaml.Node) error {
	if err := validateYAMLNodes(node); err != nil {
		return errors.New("Capture schema is invalid.")
	}
	if node.Kind != yaml.MappingNode || len(node.Content) != 2 {
		return errors.New("Capture schema is invalid.")
	}
	key, value := node.Content[0].Value, node.Content[1]
	switch key {
	case "unique-directory-entry":
		if !nonemptyYAMLString(value) {
			return errors.New("Capture schema is invalid.")
		}
		held := value.Value
		capture.UniqueDirectoryEntry = &held
	case "yaml-scalar":
		if value.Kind != yaml.MappingNode || len(value.Content) != 4 {
			return errors.New("Capture schema is invalid.")
		}
		var file, path *yaml.Node
		for index := 0; index < len(value.Content); index += 2 {
			switch value.Content[index].Value {
			case "file":
				file = value.Content[index+1]
			case "path":
				path = value.Content[index+1]
			default:
				return errors.New("Capture schema is invalid.")
			}
		}
		if file == nil || !nonemptyYAMLString(file) || path == nil || path.Kind != yaml.SequenceNode || len(path.Content) == 0 {
			return errors.New("Capture schema is invalid.")
		}
		selector := &YAMLScalarCapture{File: file.Value}
		for _, part := range path.Content {
			if !nonemptyYAMLString(part) {
				return errors.New("Capture schema is invalid.")
			}
			selector.Path = append(selector.Path, part.Value)
		}
		capture.YAMLScalar = selector
	default:
		return errors.New("Capture schema is invalid.")
	}
	return nil
}

func nonemptyYAMLString(node *yaml.Node) bool {
	return node.Kind == yaml.ScalarNode && node.ShortTag() == "!!str" && node.Value != ""
}

func validateCaptures(scenario Scenario) error {
	names := map[string]bool{}
	for _, step := range scenario.Steps {
		if step.Capture != nil && len(step.Capture) == 0 {
			return errors.New("Capture schema is invalid.")
		}
		for name, capture := range step.Capture {
			if !captureName.MatchString(name) || name == "SANDBOX" || names[name] {
				return errors.New("Capture schema is invalid.")
			}
			names[name] = true
			if len(names) > 128 {
				return errors.New("Capture schema is invalid.")
			}
			if (capture.UniqueDirectoryEntry == nil) == (capture.YAMLScalar == nil) {
				return errors.New("Capture schema is invalid.")
			}
			if capture.UniqueDirectoryEntry != nil && *capture.UniqueDirectoryEntry == "" {
				return errors.New("Capture schema is invalid.")
			}
			if capture.YAMLScalar != nil {
				if capture.YAMLScalar.File == "" || len(capture.YAMLScalar.Path) == 0 {
					return errors.New("Capture schema is invalid.")
				}
				for _, part := range capture.YAMLScalar.Path {
					if part == "" {
						return errors.New("Capture schema is invalid.")
					}
				}
			}
		}
	}
	return nil
}

func extensionEnabled(scenario Scenario) bool {
	for _, step := range scenario.Steps {
		if step.Capture != nil || strings.Contains(step.Stdin, "${") {
			return true
		}
		for _, argument := range step.Run {
			if strings.Contains(argument, "${") {
				return true
			}
		}
		for path := range step.Files {
			if strings.Contains(path, "${") {
				return true
			}
		}
		for path, expectation := range step.Expect.Files {
			if strings.Contains(path, "${") {
				return true
			}
			for _, predicate := range expectation.yamlScalars {
				if predicate.kind == "equals" || predicate.kind == "absent" {
					return true
				}
			}
		}
		for _, path := range step.Expect.Unchanged {
			if strings.Contains(path, "${") {
				return true
			}
		}
	}
	return false
}
