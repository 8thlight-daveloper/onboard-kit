package runner

import (
	"archive/zip"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"time"

	"go.yaml.in/yaml/v3"
)

type zipExpectation struct {
	entries []zipEntryExpectation
}

type zipEntryExpectation struct {
	name     string
	sameAs   string
	method   uint16
	mode     fs.FileMode
	modified time.Time
}

func decodeZIP(node *yaml.Node) (*zipExpectation, error) {
	fields, err := zipFields(node, []string{"entries"})
	if err != nil {
		return nil, err
	}
	entries := fields["entries"]
	if entries.Kind != yaml.SequenceNode || len(entries.Content) == 0 {
		return nil, errors.New("zip entries must be a nonempty sequence")
	}
	result := &zipExpectation{}
	for index, entry := range entries.Content {
		parsed, err := decodeZIPEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("zip entry %d: %w", index+1, err)
		}
		result.entries = append(result.entries, parsed)
	}
	return result, nil
}

func zipFields(node *yaml.Node, required []string) (map[string]*yaml.Node, error) {
	if node.Kind != yaml.MappingNode {
		return nil, errors.New("expected a ZIP mapping")
	}
	allowed := make(map[string]bool, len(required))
	for _, key := range required {
		allowed[key] = true
	}
	fields := make(map[string]*yaml.Node, len(required))
	for index := 0; index+1 < len(node.Content); index += 2 {
		key, value := node.Content[index], node.Content[index+1]
		if key.Kind != yaml.ScalarNode || key.ShortTag() != "!!str" || !allowed[key.Value] {
			return nil, fmt.Errorf("unknown ZIP field %q", key.Value)
		}
		if _, duplicate := fields[key.Value]; duplicate {
			return nil, fmt.Errorf("duplicate ZIP field %q", key.Value)
		}
		fields[key.Value] = value
	}
	for _, key := range required {
		if fields[key] == nil {
			return nil, fmt.Errorf("ZIP field %s is missing", key)
		}
	}
	return fields, nil
}

func decodeZIPEntry(node *yaml.Node) (zipEntryExpectation, error) {
	keys := []string{"name", "same-as", "method", "mode", "modified"}
	fields, err := zipFields(node, keys)
	if err != nil {
		return zipEntryExpectation{}, err
	}
	for _, key := range keys {
		value := fields[key]
		if value.Kind != yaml.ScalarNode || value.ShortTag() != "!!str" || value.Value == "" {
			return zipEntryExpectation{}, fmt.Errorf("ZIP %s must be nonempty text", key)
		}
	}
	result := zipEntryExpectation{name: fields["name"].Value, sameAs: fields["same-as"].Value}
	if !filepath.IsLocal(result.sameAs) {
		return zipEntryExpectation{}, errors.New("ZIP same-as must be a local repository fixture path")
	}
	switch fields["method"].Value {
	case "store":
		result.method = zip.Store
	case "deflate":
		result.method = zip.Deflate
	default:
		return zipEntryExpectation{}, errors.New("ZIP method must be store or deflate")
	}
	mode := fields["mode"].Value
	if len(mode) != 4 || mode[0] != '0' {
		return zipEntryExpectation{}, errors.New("ZIP mode must have four octal digits beginning with 0")
	}
	for _, digit := range mode {
		if digit < '0' || digit > '7' {
			return zipEntryExpectation{}, errors.New("ZIP mode must have four octal digits beginning with 0")
		}
	}
	permissions, err := strconv.ParseUint(mode, 8, 12)
	if err != nil {
		return zipEntryExpectation{}, err
	}
	result.mode = fs.FileMode(permissions)
	result.modified, err = time.Parse(time.RFC3339Nano, fields["modified"].Value)
	if err != nil {
		return zipEntryExpectation{}, fmt.Errorf("ZIP modified: %w", err)
	}
	return result, nil
}
