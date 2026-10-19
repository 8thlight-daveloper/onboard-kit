package runner

import (
	"fmt"
	"reflect"
	"sort"
)

func subsetMismatch(expected, actual any, location string) string {
	switch want := expected.(type) {
	case nil:
		if actual != nil {
			return fmt.Sprintf("%s: expected null, got %s", location, describe(actual))
		}
	case map[string]any:
		have, isObject := actual.(map[string]any)
		if !isObject {
			return fmt.Sprintf("%s: expected an object, got %s", location, describe(actual))
		}
		keys := make([]string, 0, len(want))
		for key := range want {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value, present := have[key]
			if !present {
				return fmt.Sprintf("%s.%s: missing", location, key)
			}
			if mismatch := subsetMismatch(want[key], value, location+"."+key); mismatch != "" {
				return mismatch
			}
		}
	case []any:
		have, isList := actual.([]any)
		if !isList {
			return fmt.Sprintf("%s: expected a list, got %s", location, describe(actual))
		}
		if len(have) != len(want) {
			return fmt.Sprintf("%s: expected %d items, got %d", location, len(want), len(have))
		}
		for index := range want {
			if mismatch := subsetMismatch(want[index], have[index], fmt.Sprintf("%s[%d]", location, index)); mismatch != "" {
				return mismatch
			}
		}
	default:
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Sprintf("%s: expected %s, got %s", location, describe(expected), describe(actual))
		}
	}
	return ""
}

func describe(value any) string {
	return fmt.Sprintf("%#v", value)
}
