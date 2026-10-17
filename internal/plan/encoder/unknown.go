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
		builder.add(key, c.unknownNode(fields[key], path))
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
	case int8:
		return integerNode(int64(typed))
	case int16:
		return integerNode(int64(typed))
	case int32:
		return integerNode(int64(typed))
	case int64:
		return integerNode(typed)
	case uint:
		return scalarNode(strconv.FormatUint(uint64(typed), 10))
	case uint8:
		return scalarNode(strconv.FormatUint(uint64(typed), 10))
	case uint16:
		return scalarNode(strconv.FormatUint(uint64(typed), 10))
	case uint32:
		return scalarNode(strconv.FormatUint(uint64(typed), 10))
	case uint64:
		return scalarNode(strconv.FormatUint(typed, 10))
	case float32:
		return c.decimalNode(float64(typed), 32, path)
	case float64:
		return c.decimalNode(typed, 64, path)
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

func (c *collector) decimalNode(number float64, bits int, path string) node {
	if math.IsNaN(number) || math.IsInf(number, 0) {
		c.report(codeUnsupportedValue, path, "an unknown field holds a number that cannot be written")
		return nil
	}
	text := strconv.FormatFloat(number, 'f', -1, bits)
	if !strings.Contains(text, ".") {
		text += ".0"
	}
	return scalarNode(text)
}

func (c *collector) unknownList(items []any, path string) node {
	return sequenceOf(items, func(index int, item any) node {
		return c.unknownNode(item, path+"["+strconv.Itoa(index)+"]")
	})
}

func (c *collector) unknownMap(values map[string]any, path string) node {
	builder := &mappingBuilder{}
	for _, key := range sortedKeys(values) {
		builder.add(key, c.unknownNode(values[key], path+"."+key))
	}
	return builder.build()
}
