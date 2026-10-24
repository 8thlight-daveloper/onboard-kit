package reviewer

import (
	"reflect"

	contract "onboard-kit/contracts/session/reviewer"
)

func clone[T any](value T) T { return copyValue(reflect.ValueOf(value)).Interface().(T) }

func copyValue(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(copyValue(value.Elem()))
		return result
	case reflect.Pointer:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.New(value.Type().Elem())
		result.Elem().Set(copyValue(value.Elem()))
		return result
	case reflect.Map:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeMapWithSize(value.Type(), value.Len())
		iterator := value.MapRange()
		for iterator.Next() {
			result.SetMapIndex(iterator.Key(), copyValue(iterator.Value()))
		}
		return result
	case reflect.Slice:
		if value.IsNil() {
			return reflect.Zero(value.Type())
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(copyValue(value.Index(index)))
		}
		return result
	case reflect.Struct:
		result := reflect.New(value.Type()).Elem()
		for index := 0; index < value.NumField(); index++ {
			result.Field(index).Set(copyValue(value.Field(index)))
		}
		return result
	default:
		return value
	}
}

func proposalEqual(expected, actual contract.Proposal) bool {
	return expected.NeedsEvent == actual.NeedsEvent && boundedEqual(reflect.ValueOf(expected.Source), reflect.ValueOf(actual.Source)) && boundedEqual(reflect.ValueOf(expected.Intent), reflect.ValueOf(actual.Intent)) && sessionEqual(expected.Before, actual.Before) && sessionEqual(expected.After, actual.After) && boundedEqual(reflect.ValueOf(expected.BeforeRecord), reflect.ValueOf(actual.BeforeRecord)) && boundedEqual(reflect.ValueOf(expected.AfterRecord), reflect.ValueOf(actual.AfterRecord))
}

func sessionEqual(expected, actual contract.Session) bool {
	if len(expected.Activities) == 0 && len(actual.Activities) == 0 {
		actual.Activities = expected.Activities
	}
	if len(expected.Outcomes) == 0 && len(actual.Outcomes) == 0 {
		actual.Outcomes = expected.Outcomes
	}
	if len(expected.Questions) == 0 && len(actual.Questions) == 0 {
		actual.Questions = expected.Questions
	}
	if len(expected.Blockers) == 0 && len(actual.Blockers) == 0 {
		actual.Blockers = expected.Blockers
	}
	if len(expected.Support.Problems) == 0 && len(actual.Support.Problems) == 0 {
		actual.Support.Problems = expected.Support.Problems
	}
	return boundedEqual(reflect.ValueOf(expected), reflect.ValueOf(actual))
}

func boundedEqual(expected, actual reflect.Value) bool {
	if !expected.IsValid() || !actual.IsValid() {
		return expected.IsValid() == actual.IsValid()
	}
	if expected.Type() != actual.Type() {
		return false
	}
	switch expected.Kind() {
	case reflect.Interface, reflect.Pointer:
		if expected.IsNil() || actual.IsNil() {
			return expected.IsNil() == actual.IsNil()
		}
		return boundedEqual(expected.Elem(), actual.Elem())
	case reflect.String:
		return expected.Len() == actual.Len() && expected.String() == actual.String()
	case reflect.Bool:
		return expected.Bool() == actual.Bool()
	case reflect.Int, reflect.Int64:
		return expected.Int() == actual.Int()
	case reflect.Float64:
		return expected.Float() == actual.Float()
	case reflect.Slice:
		if expected.IsNil() != actual.IsNil() || expected.Len() != actual.Len() {
			return false
		}
		for index := 0; index < expected.Len(); index++ {
			if !boundedEqual(expected.Index(index), actual.Index(index)) {
				return false
			}
		}
		return true
	case reflect.Map:
		if expected.IsNil() != actual.IsNil() || expected.Len() != actual.Len() {
			return false
		}
		iterator := expected.MapRange()
		for iterator.Next() {
			if !boundedEqual(iterator.Value(), actual.MapIndex(iterator.Key())) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for index := 0; index < expected.NumField(); index++ {
			if !boundedEqual(expected.Field(index), actual.Field(index)) {
				return false
			}
		}
		return true
	case reflect.Uint8:
		return expected.Uint() == actual.Uint()
	default:
		return false
	}
}
