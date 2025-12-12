package collections

// Apply applies the applicator function to each item in the input slice.
func Apply[T, V any](items []T, applicator func(T) V) []V {
	result := make([]V, len(items))
	for i, item := range items {
		result[i] = applicator(item)
	}
	return result
}

func ApplyVariadic[T, V any](applicator func(T) V, items ...T) []V {
	return Apply(items, applicator)
}

// SliceFromVariadic creates a slice from variadic arguments.
func SliceFromVariadic[T any](items ...T) []T {
	return items
}
