package sliceale

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestIsOneOf(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		// Given
		s := make([]int, 0)

		// When
		result := IsOneOf(s, 1)

		// Then
		require.Equal(t, false, result)
	})

	t.Run("element is in the slice", func(t *testing.T) {
		// Given
		s := []int{1, 2, 3}

		// When
		result := IsOneOf(s, 2)

		// Then
		require.Equal(t, true, result)
	})

	t.Run("element is in the slice multiple times", func(t *testing.T) {
		// Given
		s := []int{1, 2, 3, 4, 5, 6, 7, 2, 2, 6, 7, 8, 9, 2}

		// When
		result := IsOneOf(s, 2)

		// Then
		require.Equal(t, true, result)
	})

	t.Run("element is not in the slice", func(t *testing.T) {
		// Given
		s := []int{1, 2, 3}

		// When
		result := IsOneOf(s, 4)

		// Then
		require.Equal(t, false, result)
	})

	t.Run("element is in the slice with multiple types", func(t *testing.T) {
		// Given
		s := []interface{}{1, "2", 3}

		// When
		result := IsOneOf(s, "2")

		// Then
		require.Equal(t, true, result)
	})

	t.Run("element is not in the slice with multiple types", func(t *testing.T) {
		// Given
		s := []interface{}{1, "2", 3}

		// When
		result := IsOneOf(s, 2)

		// Then
		require.Equal(t, false, result)
	})

}
