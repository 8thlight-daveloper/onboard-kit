package loader

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
	var documents []yaml.Node
	for {
		var document yaml.Node
		err := stream.Decode(&document)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, codeNotYAML
		}
		documents = append(documents, document)
	}
	if len(documents) == 0 || isEmptyDocument(documents[0]) {
		return nil, codeNotATemplate
	}
	if len(documents) > 1 {
		return nil, codeMultipleDocuments
	}
	root := documents[0].Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, codeNotATemplate
	}
	return root, ""
}

func isEmptyDocument(document yaml.Node) bool {
	content := document.Content[0]
	return content.Kind == yaml.ScalarNode && content.Value == "" && content.Style&quotedOrBlockStyle == 0 && content.Anchor == ""
}
