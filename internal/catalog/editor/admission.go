package editor

import (
	"math"
	contract "onboard-kit/contracts/catalog/editor"
	"onboard-kit/internal/model"
	"reflect"
	"sort"
	"unicode/utf8"
)

type admission struct {
	entries int
	text    int
	active  map[uintptr]bool
	code    string
}

func admit(plan model.Plan, operations []contract.Operation) (model.Plan, []contract.Operation, string) {
	if len(operations) > 1024 || len(plan.Sections) > 4096 {
		return model.Plan{}, nil, "edit.limit"
	}
	adds := 0
	for _, operation := range operations {
		if operation.Kind == contract.Add {
			adds++
		}
	}
	if adds > 256 {
		return model.Plan{}, nil, "edit.limit"
	}
	checking := admission{active: make(map[uintptr]bool)}
	checking.walk(reflect.ValueOf(plan), -1)
	checking.walk(reflect.ValueOf(operations), -1)
	if checking.code != "" {
		return model.Plan{}, nil, checking.code
	}
	return clone(reflect.ValueOf(plan)).Interface().(model.Plan), clone(reflect.ValueOf(operations)).Interface().([]contract.Operation), ""
}

func (checking *admission) walk(value reflect.Value, depth int) {
	if checking.code != "" || !value.IsValid() {
		return
	}
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return
		}
		checking.walk(value.Elem(), depth)
		return
	}
	if depth < 0 && (value.Type() == reflect.TypeFor[map[string]any]() || value.Type() == reflect.TypeFor[[]any]()) {
		depth = 0
	}
	if depth >= 0 {
		switch value.Interface().(type) {
		case map[string]any, []any, string, bool, int, int64, nil:
		case float64:
			if math.IsNaN(value.Float()) || math.IsInf(value.Float(), 0) {
				checking.code = "edit.held-shape"
				return
			}
		default:
			checking.code = "edit.held-shape"
			return
		}
	}
	switch value.Kind() {
	case reflect.String:
		checking.string(value.String(), 1048576)
	case reflect.Pointer:
		if !value.IsNil() {
			checking.walk(value.Elem(), depth)
		}
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			field := value.Field(index)
			name := value.Type().Field(index).Name
			operation := value.Type() == reflect.TypeFor[contract.Operation]()
			binding := value.Type() == reflect.TypeFor[contract.Binding]()
			if (operation && (name == "SectionID" || name == "SourceID" || name == "Name")) || (binding && name == "Name") {
				checking.string(field.String(), 4096)
			} else {
				checking.walk(field, depth)
			}
			if checking.code != "" {
				return
			}
		}
	case reflect.Slice, reflect.Map:
		checking.collection(value, depth)
	}
}

func (checking *admission) string(value string, limit int) {
	if checking.code != "" {
		return
	}
	if len(value) > limit || len(value) > 33554432-checking.text {
		checking.code = "edit.limit"
		return
	}
	checking.text += len(value)
	if !utf8.ValidString(value) {
		checking.code = "edit.text"
	}
}

func (checking *admission) collection(value reflect.Value, depth int) {
	if value.Len() > 8192 || value.Len() > 262144-checking.entries {
		checking.code = "edit.limit"
		return
	}
	checking.entries += value.Len()
	childDepth := depth
	if depth >= 0 {
		if depth > 64 {
			checking.code = "edit.limit"
			return
		}
		childDepth++
		if !value.IsNil() && value.Len() != 0 {
			identity := uintptr(value.UnsafePointer())
			if checking.active[identity] {
				checking.code = "edit.limit"
				return
			}
			checking.active[identity] = true
			defer delete(checking.active, identity)
		}
	}
	if value.Kind() == reflect.Slice {
		for index := 0; index < value.Len(); index++ {
			checking.walk(value.Index(index), childDepth)
			if checking.code != "" {
				return
			}
		}
		return
	}
	keys := value.MapKeys()
	sort.Slice(keys, func(left, right int) bool { return keys[left].String() < keys[right].String() })
	for _, key := range keys {
		checking.string(key.String(), 1048576)
		checking.walk(value.MapIndex(key), childDepth)
		if checking.code != "" {
			return
		}
	}
}

func clone(value reflect.Value) reflect.Value {
	if !value.IsValid() {
		return value
	}
	result := reflect.New(value.Type()).Elem()
	switch value.Kind() {
	case reflect.Interface:
		if !value.IsNil() {
			result.Set(clone(value.Elem()))
		}
	case reflect.Pointer:
		if !value.IsNil() {
			result.Set(reflect.New(value.Type().Elem()))
			result.Elem().Set(clone(value.Elem()))
		}
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			result.Field(index).Set(clone(value.Field(index)))
		}
	case reflect.Slice:
		if !value.IsNil() {
			result.Set(reflect.MakeSlice(value.Type(), value.Len(), value.Len()))
			for index := 0; index < value.Len(); index++ {
				result.Index(index).Set(clone(value.Index(index)))
			}
		}
	case reflect.Map:
		if !value.IsNil() {
			result.Set(reflect.MakeMapWithSize(value.Type(), value.Len()))
			iterator := value.MapRange()
			for iterator.Next() {
				result.SetMapIndex(iterator.Key(), clone(iterator.Value()))
			}
		}
	default:
		result.Set(value)
	}
	return result
}
