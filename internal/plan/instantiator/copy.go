package instantiator

func copySlice[Item any](items []Item, copyItem func(Item) Item) []Item {
	if items == nil {
		return nil
	}
	copied := make([]Item, len(items))
	for index, item := range items {
		copied[index] = copyItem(item)
	}
	return copied
}

func copyStrings(items []string) []string {
	return copySlice(items, func(item string) string { return item })
}

func copyStringMap(entries map[string]string) map[string]string {
	if entries == nil {
		return nil
	}
	copied := make(map[string]string, len(entries))
	for key, entry := range entries {
		copied[key] = entry
	}
	return copied
}
