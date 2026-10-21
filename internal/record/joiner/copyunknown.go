package joiner

func copyUnknownFields(fields map[string]any) map[string]any {
	return copyMap(fields, copyUnknown)
}

func copyUnknown(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return copyUnknownFields(typed)
	case []any:
		return copyEach(typed, copyUnknown)
	default:
		return value
	}
}
