package decoder

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"

	"go.yaml.in/yaml/v3"
)

func parseDocument(data []byte) (*yaml.Node, string) {
	if !utf8.Valid(data) {
		return nil, codeNotYAML
	}
	stream := yaml.NewDecoder(bytes.NewReader(data))
	var first yaml.Node
	if err := stream.Decode(&first); err != nil {
		return nil, firstDocumentCode(err)
	}
	var second yaml.Node
	switch err := stream.Decode(&second); {
	case err == nil:
		return nil, codeMultipleDocuments
	case !errors.Is(err, io.EOF):
		return nil, codeNotYAML
	}
	root := first.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, codeNotAPlan
	}
	return root, ""
}

func firstDocumentCode(err error) string {
	if errors.Is(err, io.EOF) {
		return codeNotAPlan
	}
	return codeNotYAML
}
