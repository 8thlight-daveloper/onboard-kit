package joiner

func identity[T any](item T) T {
	return item
}

func copyEach[T any](items []T, copyItem func(T) T) []T {
	if items == nil {
		return nil
	}
	copied := make([]T, len(items))
	for index, item := range items {
		copied[index] = copyItem(item)
	}
	return copied
}

func copyMap[T any](entries map[string]T, copyItem func(T) T) map[string]T {
	if entries == nil {
		return nil
	}
	copied := make(map[string]T, len(entries))
	for key, item := range entries {
		copied[key] = copyItem(item)
	}
	return copied
}

func copyPointer[T any](pointer *T) *T {
	if pointer == nil {
		return nil
	}
	copied := *pointer
	return &copied
}
