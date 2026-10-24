package runner

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"time"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

const yamlScalarByteLimit int64 = 1048576

var utcTimestampShape = regexp.MustCompile(`^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]{1,9})?Z$`)

func checkYAMLScalarFile(path string, predicates []yamlScalarExpectation, folder string) []Failure {
	label := "file " + path
	target, err := resolve(folder, path)
	if err != nil {
		return []Failure{{Expectation: label, Message: "file path must stay inside the scenario folder"}}
	}
	content, reason, fileFailure := readYAMLScalarRegular(target)
	if reason != "" {
		if !fileFailure {
			label += " yaml-scalars"
		}
		return []Failure{{Expectation: label, Message: reason}}
	}
	label += " yaml-scalars"
	root, err := decodeYAMLScalarDocument(content)
	if err != nil {
		return []Failure{{Expectation: label, Message: err.Error()}}
	}
	var failures []Failure
	for index, predicate := range predicates {
		if reason := yamlPredicateMismatch(root, predicate); reason != "" {
			failures = append(failures, Failure{Expectation: label, Message: fmt.Sprintf("predicate %d: %s", index+1, reason)})
		}
	}
	return failures
}

func readYAMLScalarRegular(path string) ([]byte, string, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, "the file could not be opened", true
	}
	if !info.Mode().IsRegular() {
		return nil, "expected a regular file", true
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, "the file could not be opened", true
	}
	info, err = file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		file.Close()
		return nil, "expected an accessible regular file", true
	}
	content, readErr := io.ReadAll(io.LimitReader(file, yamlScalarByteLimit+1))
	closeErr := file.Close()
	if readErr != nil || closeErr != nil {
		return nil, "the file could not be read", true
	}
	if int64(len(content)) > yamlScalarByteLimit {
		return nil, "YAML file exceeds the 1048576 byte limit", false
	}
	return content, "", false
}

func decodeYAMLScalarDocument(content []byte) (*yaml.Node, error) {
	if !utf8.Valid(content) {
		return nil, errors.New("YAML file is not valid UTF8")
	}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	var document yaml.Node
	if err := decoder.Decode(&document); err != nil {
		return nil, errors.New("invalid or empty YAML document")
	}
	if document.Kind != yaml.DocumentNode || len(document.Content) != 1 {
		return nil, errors.New("expected one nonempty YAML document")
	}
	root := document.Content[0]
	if root.Kind == yaml.ScalarNode && root.ShortTag() == "!!null" && root.Value == "" {
		return nil, errors.New("expected one nonempty YAML document")
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return nil, errors.New("expected exactly one YAML document")
	}
	if err := validateYAMLNodes(&document); err != nil {
		return nil, err
	}
	return root, nil
}

func yamlScalarMismatch(root *yaml.Node, path []string) string {
	return yamlPredicateMismatch(root, yamlScalarExpectation{path: path, kind: "utc-timestamp"})
}

func selectYAML(root *yaml.Node, path []string) (*yaml.Node, bool, error) {
	selected := root
	for _, part := range path {
		switch selected.Kind {
		case yaml.MappingNode:
			var next *yaml.Node
			for index := 0; index+1 < len(selected.Content); index += 2 {
				if selected.Content[index].Value == part {
					next = selected.Content[index+1]
					break
				}
			}
			if next == nil {
				return nil, false, nil
			}
			selected = next
		case yaml.SequenceNode:
			if part == "" || (len(part) > 1 && part[0] == '0') {
				return nil, false, errors.New("invalid sequence index")
			}
			for _, character := range part {
				if character < '0' || character > '9' {
					return nil, false, errors.New("invalid sequence index")
				}
			}
			index, err := strconv.ParseUint(part, 10, 63)
			if err != nil {
				return nil, false, errors.New("invalid sequence index")
			}
			if index >= uint64(len(selected.Content)) {
				return nil, false, nil
			}
			selected = selected.Content[int(index)]
		default:
			return nil, false, errors.New("invalid path traversal")
		}
	}
	return selected, true, nil
}

func yamlPredicateMismatch(root *yaml.Node, predicate yamlScalarExpectation) string {
	selected, present, err := selectYAML(root, predicate.path)
	if err != nil {
		return err.Error()
	}
	if predicate.kind == "absent" {
		if !present {
			return ""
		}
		return "path is present"
	}
	if !present {
		return "missing path"
	}
	if selected.Kind != yaml.ScalarNode || selected.ShortTag() != "!!str" {
		return "selected value must be a YAML string scalar"
	}
	if predicate.kind == "equals" {
		if selected.Value == predicate.equals {
			return ""
		}
		return "selected value differs"
	}
	if !utcTimestampShape.MatchString(selected.Value) || selected.Value[:4] == "0000" {
		return "invalid UTC timestamp"
	}
	if _, err := time.Parse(time.RFC3339Nano, selected.Value); err != nil {
		return "invalid UTC timestamp"
	}
	return ""
}
