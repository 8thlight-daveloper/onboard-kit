package valuejson

import (
	"errors"
	"fmt"

	"onboard-kit/internal/model"
)

func Decode(raw any) (model.Value, error) {
	switch typed := raw.(type) {
	case nil:
		return model.Value{}, nil
	case string:
		return model.Value{Kind: model.ValueKindText, Text: typed}, nil
	case bool:
		return model.Value{Kind: model.ValueKindFlag, Flag: typed}, nil
	case []any:
		return decodeList(typed)
	case map[string]any:
		return decodeObject(typed)
	}
	return model.Value{}, fmt.Errorf("a value is text, true or false, a list of text, or {\"deferred\": true}")
}

func decodeList(items []any) (model.Value, error) {
	list := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return model.Value{}, errors.New("a list holds only text")
		}
		list = append(list, text)
	}
	return model.Value{Kind: model.ValueKindList, List: list}, nil
}

func decodeObject(object map[string]any) (model.Value, error) {
	if flag, ok := object["deferred"].(bool); ok && flag && len(object) == 1 {
		return model.Value{Kind: model.ValueKindDeferred}, nil
	}
	return model.Value{}, errors.New("the only object a value may be is {\"deferred\": true}")
}

func Encode(value model.Value) any {
	switch value.Kind {
	case model.ValueKindText:
		return value.Text
	case model.ValueKindFlag:
		return value.Flag
	case model.ValueKindList:
		if value.List == nil {
			return []string{}
		}
		return value.List
	case model.ValueKindDeferred:
		return map[string]any{"deferred": true}
	}
	return nil
}
