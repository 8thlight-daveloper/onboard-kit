package encoder

import (
	"math"
	"sort"
	"strconv"
	"strings"
)

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (c *collector) addUnknownFields(builder *mappingBuilder, fields map[string]any, prefix string, isKnown func(string) bool) {
	for _, key := range sortedKeys(fields) {
		path := prefix + key
		if isKnown(key) {
			c.report(codeUnknownFieldCollision, path, "an unknown field has the key of a known field")
			continue
		}
		builder.addNamed(key, c.unknownNode(fields[key], path))
	}
}

func (c *collector) unknownNode(value any, path string) node {
	switch typed := value.(type) {
	case nil:
		return scalarNode("null")
	case string:
		return textNode(typed)
	case bool:
		return booleanNode(typed)
	case int:
		return integerNode(int64(typed))
	case int64:
		return integerNode(typed)
	case float64:
		return c.decimalNode(typed, path)
	case []any:
		return c.unknownList(typed, path)
	case map[string]any:
		return c.unknownMap(typed, path)
	}
	c.report(codeUnsupportedValue, path, "an unknown field holds a value that cannot be written")
	return nil
}

func integerNode(number int64) node {
	return scalarNode(strconv.FormatInt(number, 10))
}

func (c *collector) decimalNode(number float64, path string) node {
	if math.IsNaN(number) || math.IsInf(number, 0) {
		c.report(codeUnsupportedValue, path, "an unknown field holds a number that cannot be written")
		return nil
	}
	text := strconv.FormatFloat(number, 'f', -1, 64)
	if !strings.Contains(text, ".") {
		text += ".0"
	}
	return scalarNode(text)
}

func (c *collector) unknownList(items []any, path string) node {
	if len(items) == 0 {
		return scalarNode("[]")
	}
	return sequenceOf(items, func(index int, item any) node {
		return c.unknownNode(item, indexed(path, index))
	})
}

func (c *collector) unknownMap(values map[string]any, path string) node {
	if len(values) == 0 {
		return scalarNode("{}")
	}
	builder := &mappingBuilder{}
	for _, key := range sortedKeys(values) {
		builder.addNamed(key, c.unknownNode(values[key], path+"."+key))
	}
	return builder.build()
}
