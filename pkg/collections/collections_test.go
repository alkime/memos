package collections_test

import (
	"testing"

	"github.com/alkime/memos/pkg/collections"

	"github.com/stretchr/testify/require"
)

func TestApply(t *testing.T) {
	t.Run("basic types", func(t *testing.T) {
		ints := []int{1, 2, 3, 4}
		squared := collections.Apply(ints, func(i int) int {
			return i * i
		})

		expected := []int{1, 4, 9, 16}
		require.ElementsMatch(t, expected, squared)

		strs := []string{"a", "bb", "ccc"}
		lengths := collections.Apply(strs, func(s string) int {
			return len(s)
		})

		expectedLengths := []int{1, 2, 3}
		require.ElementsMatch(t, expectedLengths, lengths)
	})

	t.Run("structs", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}

		people := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
			{Name: "Charlie", Age: 35},
		}

		names := collections.Apply(people, func(p Person) string {
			return p.Name
		})

		expectedNames := []string{"Alice", "Bob", "Charlie"}
		require.ElementsMatch(t, expectedNames, names)

		expectedAges := []int{30, 25, 35}
		ages := collections.Apply(people, func(p Person) int {
			return p.Age
		})
		require.ElementsMatch(t, expectedAges, ages)
	})
}
