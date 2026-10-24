package runner

import (
	"errors"
	"strings"
)

const expandedFieldLimit = 1048576

func expandField(input string, values map[string]string) (string, error) {
	var result strings.Builder
	for position := 0; position < len(input); {
		next := strings.Index(input[position:], "${")
		if next < 0 {
			if result.Len()+len(input)-position > expandedFieldLimit {
				return "", errors.New("interpolation")
			}
			result.WriteString(input[position:])
			break
		}
		next += position
		if result.Len()+next-position > expandedFieldLimit {
			return "", errors.New("interpolation")
		}
		result.WriteString(input[position:next])
		end := strings.IndexByte(input[next+2:], '}')
		if end < 0 {
			return "", errors.New("interpolation")
		}
		end += next + 2
		name := input[next+2 : end]
		value, known := values[name]
		if !captureName.MatchString(name) || !known || result.Len()+len(value) > expandedFieldLimit {
			return "", errors.New("interpolation")
		}
		result.WriteString(value)
		position = end + 1
	}
	return result.String(), nil
}

func expandStep(original Step, values map[string]string) (Step, error) {
	step := original
	if len(original.Run) == 0 || strings.Contains(original.Run[0], "${") {
		return Step{}, errors.New("interpolation")
	}
	step.Run = append([]string(nil), original.Run...)
	var err error
	for index := 1; index < len(step.Run); index++ {
		step.Run[index], err = expandField(step.Run[index], values)
		if err != nil {
			return Step{}, err
		}
	}
	step.Stdin, err = expandField(original.Stdin, values)
	if err != nil {
		return Step{}, err
	}
	step.Files = make(map[string]FileSource, len(original.Files))
	for path, source := range original.Files {
		expanded, err := expandField(path, values)
		if err != nil {
			return Step{}, err
		}
		if _, exists := step.Files[expanded]; exists {
			return Step{}, errors.New("interpolation")
		}
		step.Files[expanded] = source
	}
	step.Expect = original.Expect
	step.Expect.Files = make(map[string]FileExpectation, len(original.Expect.Files))
	for path, expectation := range original.Expect.Files {
		expanded, err := expandField(path, values)
		if err != nil {
			return Step{}, err
		}
		if _, exists := step.Expect.Files[expanded]; exists {
			return Step{}, errors.New("interpolation")
		}
		expectation.yamlScalars = append([]yamlScalarExpectation(nil), expectation.yamlScalars...)
		for index := range expectation.yamlScalars {
			if expectation.yamlScalars[index].kind == "equals" {
				expectation.yamlScalars[index].equals, err = expandField(expectation.yamlScalars[index].equals, values)
				if err != nil {
					return Step{}, err
				}
			}
		}
		step.Expect.Files[expanded] = expectation
	}
	step.Expect.Unchanged = make([]string, len(original.Expect.Unchanged))
	seen := map[string]bool{}
	for index, path := range original.Expect.Unchanged {
		expanded, err := expandField(path, values)
		if err != nil {
			return Step{}, err
		}
		if seen[expanded] {
			return Step{}, errors.New("interpolation")
		}
		seen[expanded] = true
		step.Expect.Unchanged[index] = expanded
	}
	step.Capture = make(map[string]Capture, len(original.Capture))
	for name, capture := range original.Capture {
		if capture.UniqueDirectoryEntry != nil {
			expanded, err := expandField(*capture.UniqueDirectoryEntry, values)
			if err != nil {
				return Step{}, err
			}
			capture.UniqueDirectoryEntry = &expanded
		}
		if capture.YAMLScalar != nil {
			selector := *capture.YAMLScalar
			selector.File, err = expandField(selector.File, values)
			if err != nil {
				return Step{}, err
			}
			selector.Path = append([]string(nil), selector.Path...)
			capture.YAMLScalar = &selector
		}
		step.Capture[name] = capture
	}
	return step, nil
}
