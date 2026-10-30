package editguide

import (
	"math"
	"reflect"
)

func equivalent(left, right any) bool {
	return sameValue(reflect.ValueOf(left), reflect.ValueOf(right))
}

func plainInteger(kind reflect.Kind) bool { return kind == reflect.Int || kind == reflect.Int64 }

func sameValue(left, right reflect.Value) bool {
	if left.Kind() == reflect.Interface || right.Kind() == reflect.Interface {
		if left.Kind() != right.Kind() || left.IsNil() || right.IsNil() {
			return left.Kind() == right.Kind() && left.IsNil() == right.IsNil()
		}
		return sameValue(left.Elem(), right.Elem())
	}
	if plainInteger(left.Kind()) && plainInteger(right.Kind()) {
		return left.Int() == right.Int()
	}
	if left.Type() != right.Type() {
		return false
	}
	switch left.Kind() {
	case reflect.Struct:
		for i := 0; i < left.NumField(); i++ {
			if !sameValue(left.Field(i), right.Field(i)) {
				return false
			}
		}
		return true
	case reflect.Slice, reflect.Array:
		if left.Len() != right.Len() {
			return false
		}
		for i := 0; i < left.Len(); i++ {
			if !sameValue(left.Index(i), right.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if left.Len() != right.Len() {
			return false
		}
		for _, key := range left.MapKeys() {
			other := right.MapIndex(key)
			if !other.IsValid() || !sameValue(left.MapIndex(key), other) {
				return false
			}
		}
		return true
	case reflect.Pointer:
		if left.IsNil() || right.IsNil() {
			return left.IsNil() == right.IsNil()
		}
		return sameValue(left.Elem(), right.Elem())
	case reflect.String:
		return left.String() == right.String()
	case reflect.Bool:
		return left.Bool() == right.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return left.Int() == right.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return left.Uint() == right.Uint()
	case reflect.Float32, reflect.Float64:
		return left.Float() == right.Float() || math.IsNaN(left.Float()) && math.IsNaN(right.Float())
	}
	return false
}
