package decoder

import (
	"regexp"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

type scalarKind int

const (
	scalarText scalarKind = iota
	scalarNull
	scalarBoolean
	scalarBooleanSpelling
	scalarNullSpelling
	scalarInteger
	scalarDecimal
	scalarNumberSpelling
)

const (
	quotedOrBlockStyle = yaml.DoubleQuotedStyle | yaml.SingleQuotedStyle | yaml.LiteralStyle | yaml.FoldedStyle
	deferredText       = "deferred"
)

var (
	canonicalDecimal = regexp.MustCompile(`^-?(0|[1-9][0-9]*)\.[0-9]+$`)
	numberSpellings  = []*regexp.Regexp{
		regexp.MustCompile(`^[+-]?0[0-9]+(\.[0-9]+)?$`),
		regexp.MustCompile(`^\+[0-9]+(\.[0-9]+)?$`),
		regexp.MustCompile(`^[+-]?\.[0-9]+$`),
		regexp.MustCompile(`^[+-]?[0-9]+\.$`),
		regexp.MustCompile(`^[+-]?[0-9]+(_[0-9]+)*(\.[0-9]+(_[0-9]+)*)?$`),
		regexp.MustCompile(`^[+-]?([0-9]+(_[0-9]+)*\.?([0-9]+(_[0-9]+)*)?|\.[0-9]+(_[0-9]+)*)[eE][+-]?[0-9]+$`),
		regexp.MustCompile(`^[+-]?0([xX][0-9a-fA-F]+(_[0-9a-fA-F]+)*|[oO][0-7]+(_[0-7]+)*|[bB][01]+(_[01]+)*)$`),
		regexp.MustCompile(`^(?i)[+-]?\.(inf|nan)$`),
	}
	booleanWords = map[string]bool{
		"yes": true, "no": true, "on": true, "off": true, "y": true, "n": true, "true": true, "false": true,
	}
)

func isPlain(node *yaml.Node) bool {
	return node.Kind == yaml.ScalarNode && node.Style&quotedOrBlockStyle == 0
}

func classify(node *yaml.Node) scalarKind {
	text := node.Value
	switch {
	case text == "" || text == "null":
		return scalarNull
	case text == "~" || text == "Null" || text == "NULL":
		return scalarNullSpelling
	case text == "true" || text == "false":
		return scalarBoolean
	case booleanWords[strings.ToLower(text)]:
		return scalarBooleanSpelling
	}
	return classifyNumber(text)
}

func classifyNumber(text string) scalarKind {
	if canonicalInteger.MatchString(text) {
		if _, err := strconv.ParseInt(text, 10, 64); err != nil {
			return scalarNumberSpelling
		}
		return scalarInteger
	}
	if canonicalDecimal.MatchString(text) {
		if _, err := strconv.ParseFloat(text, 64); err != nil {
			return scalarNumberSpelling
		}
		return scalarDecimal
	}
	for _, spelling := range numberSpellings {
		if spelling.MatchString(text) {
			return scalarNumberSpelling
		}
	}
	return scalarText
}

func isNull(node *yaml.Node) bool {
	return isPlain(node) && classify(node) == scalarNull
}

func isCanonicalNumber(kind scalarKind) bool {
	return kind == scalarInteger || kind == scalarDecimal
}
