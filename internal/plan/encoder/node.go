package encoder

import "strings"

type node interface{}

type scalarNode string

type literalNode struct {
	chomp string
	lines []string
}

type entry struct {
	key   string
	value node
}

type mappingNode []entry

type sequenceNode []node

type mappingBuilder struct {
	entries mappingNode
}

func (b *mappingBuilder) add(key string, value node) {
	if value != nil {
		b.entries = append(b.entries, entry{key: keyText(key), value: value})
	}
}

func (b *mappingBuilder) build() node {
	if len(b.entries) == 0 {
		return nil
	}
	return b.entries
}

func sequenceOf[T any](items []T, build func(int, T) node) node {
	var nodes sequenceNode
	for index, item := range items {
		if built := build(index, item); built != nil {
			nodes = append(nodes, built)
		}
	}
	if len(nodes) == 0 {
		return nil
	}
	return nodes
}

func padding(indent int) string {
	return strings.Repeat(" ", indent)
}

func render(document node, indent int) []string {
	switch block := document.(type) {
	case mappingNode:
		return renderMapping(block, indent)
	case sequenceNode:
		return renderSequence(block, indent)
	}
	return nil
}

func renderMapping(mapping mappingNode, indent int) []string {
	var lines []string
	for _, item := range mapping {
		lines = append(lines, renderEntry(item, indent)...)
	}
	return lines
}

func renderEntry(item entry, indent int) []string {
	head := padding(indent) + item.key + ":"
	switch value := item.value.(type) {
	case scalarNode:
		return []string{head + " " + string(value)}
	case literalNode:
		return append([]string{head + " |" + value.chomp}, literalLines(value, indent+2)...)
	}
	return append([]string{head}, render(item.value, indent+2)...)
}

func renderSequence(sequence sequenceNode, indent int) []string {
	var lines []string
	for _, item := range sequence {
		lines = append(lines, renderItem(item, indent)...)
	}
	return lines
}

func renderItem(item node, indent int) []string {
	switch value := item.(type) {
	case scalarNode:
		return []string{padding(indent) + "- " + string(value)}
	case literalNode:
		return append([]string{padding(indent) + "- |" + value.chomp}, literalLines(value, indent+2)...)
	}
	lines := render(item, indent+2)
	lines[0] = padding(indent) + "- " + lines[0][indent+2:]
	return lines
}

func literalLines(literal literalNode, indent int) []string {
	lines := make([]string, len(literal.lines))
	for index, line := range literal.lines {
		if line != "" {
			lines[index] = padding(indent) + line
		}
	}
	return lines
}
