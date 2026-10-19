package instantiator

func copyUnknownMap(entries map[string]any) map[string]any {
	if entries == nil {
		return nil
	}
	copied := make(map[string]any, len(entries))
	for key, entry := range entries {
		copied[key] = copyUnknown(entry)
	}
	return copied
}

func copyUnknown(value any) any {
	switch typed := value.(type) {
	case []any:
		return copySlice(typed, copyUnknown)
	case []string:
		return copyStrings(typed)
	case map[string]any:
		return copyUnknownMap(typed)
	case map[string]string:
		return copyStringMap(typed)
	default:
		return value
	}
}
