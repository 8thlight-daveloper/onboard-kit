package decoder

import "go.yaml.in/yaml/v3"

const mergeKey = "<<"

type entry struct {
	name  string
	key   *yaml.Node
	value *yaml.Node
	at    location
}

func (c *collector) open(node *yaml.Node, at location) bool {
	if node.Kind == yaml.AliasNode || node.Anchor != "" {
		c.report(codeAlias, at)
	}
	return node.Kind != yaml.AliasNode
}

func (c *collector) prepare(node *yaml.Node, at location, canonicalNumbersAreSpelling bool) bool {
	if !c.open(node, at) {
		return false
	}
	if !isPlain(node) {
		return true
	}
	code := spellingCode(classify(node), canonicalNumbersAreSpelling)
	if code == "" {
		return true
	}
	c.report(code, at)
	return false
}

func spellingCode(kind scalarKind, canonicalNumbersAreSpelling bool) string {
	switch {
	case kind == scalarBooleanSpelling:
		return codeBooleanSpelling
	case kind == scalarNullSpelling:
		return codeNullSpelling
	case kind == scalarNumberSpelling:
		return codeNumberSpelling
	case canonicalNumbersAreSpelling && isCanonicalNumber(kind):
		return codeNumberSpelling
	}
	return ""
}

func (c *collector) entries(mapping *yaml.Node, at location) []entry {
	var found []entry
	seen := map[string]bool{}
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		key, value := mapping.Content[index], mapping.Content[index+1]
		if isPlain(key) && key.Value == mergeKey {
			c.report(codeAlias, at.field(mergeKey, key))
			continue
		}
		place := at.field(key.Value, key)
		c.open(key, place)
		if seen[key.Value] {
			c.report(codeDuplicateKey, place)
		}
		seen[key.Value] = true
		found = append(found, entry{name: key.Value, key: key, value: value, at: place})
	}
	return found
}

func (c *collector) mappingOf(node *yaml.Node, at location) *yaml.Node {
	if !c.prepare(node, at, false) || isNull(node) {
		return nil
	}
	if node.Kind != yaml.MappingNode {
		c.report(codeWrongType, at)
		return nil
	}
	return node
}

func (c *collector) rejectUnknown(found entry) {
	c.report(codeUnknownField, found.at)
}

func (c *collector) items(node *yaml.Node, at location, nullCode string, read func(*yaml.Node, location)) {
	if !c.prepare(node, at, false) || isNull(node) {
		return
	}
	if node.Kind != yaml.SequenceNode {
		c.report(codeWrongType, at)
		return
	}
	c.eachItem(node, at, nullCode, read)
}

func (c *collector) eachItem(sequence *yaml.Node, at location, nullCode string, read func(*yaml.Node, location)) {
	for index, item := range sequence.Content {
		place := at.item(index, item)
		if isNull(item) {
			c.report(nullCode, place)
			continue
		}
		read(item, place)
	}
}
