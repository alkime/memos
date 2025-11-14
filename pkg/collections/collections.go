package collections

// Apply applies the applicator function to each item in the input slice.
func Apply[T, V any](items []T, applicator func(T) V) []V {
	result := make([]V, len(items))
	for i, item := range items {
		result[i] = applicator(item)
	}
	return result
}
