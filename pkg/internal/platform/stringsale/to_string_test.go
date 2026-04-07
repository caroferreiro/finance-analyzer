package stringsale

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveDuplicateSpaces(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		// Given
		input := ""

		// When
		result := RemoveDuplicateSpaces(input)

		// Then
		require.Empty(t, result)
	})

	t.Run("string with no spaces", func(t *testing.T) {
		// Given
		input := "hello"

		// When
		result := RemoveDuplicateSpaces(input)

		// Then
		expected := "hello"
		require.Equal(t, expected, result)
	})

	t.Run("string with single space", func(t *testing.T) {
		// Given
		input := "hello world"

		// When
		result := RemoveDuplicateSpaces(input)

		// Then
		expected := "hello world"
		require.Equal(t, expected, result)
	})

	t.Run("string with multiple spaces", func(t *testing.T) {
		// Given
		input := "hello     world"

		// When
		result := RemoveDuplicateSpaces(input)

		// Then
		expected := "hello world"
		require.Equal(t, expected, result)

		// Input wasn't modified
		require.Equal(t, "hello     world", input)
	})
}
