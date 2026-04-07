package pdftable

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFakeTableIterator_Next(t *testing.T) {
	t.Run("Empty rows", func(t *testing.T) {
		it := NewFakeTableIterator([]Row{})
		row, ok := it.Next()
		require.False(t, ok)
		require.Empty(t, row)
	})

	t.Run("Single row", func(t *testing.T) {
		row1 := Row{RawText: "row1"}
		it := NewFakeTableIterator([]Row{row1})
		row, ok := it.Next()
		require.True(t, ok)
		require.Equal(t, row1, row)
		row, ok = it.Next()
		require.False(t, ok)
		require.Empty(t, row)
	})

	t.Run("Multiple rows", func(t *testing.T) {
		row1 := Row{RawText: "row1"}
		row2 := Row{RawText: "row2"}
		it := NewFakeTableIterator([]Row{row1, row2})
		row, ok := it.Next()
		require.True(t, ok)
		require.Equal(t, row1, row)
		row, ok = it.Next()
		require.True(t, ok)
		require.Equal(t, row2, row)
		row, ok = it.Next()
		require.False(t, ok)
		require.Empty(t, row)
	})
}

func TestFakeTableIterator_NextUtilRegexIsMatched(t *testing.T) {
	row1 := Row{RawText: "foo bar"}
	row2 := Row{RawText: "baz qux"}
	row3 := Row{RawText: "hello world"}
	it := NewFakeTableIterator([]Row{row1, row2, row3})

	// Regex that matches 'baz'
	regex := regexp.MustCompile(`baz`)
	row, ok := it.NextUtilRegexIsMatched(regex)
	require.True(t, ok)
	require.Equal(t, row2, row)

	// Regex that matches nothing in the remaining rows
	regexNone := regexp.MustCompile(`notfound`)
	row, ok = it.NextUtilRegexIsMatched(regexNone)
	require.False(t, ok)
	require.Empty(t, row)

	// Regex that matches 'hello' in the last row
	it = NewFakeTableIterator([]Row{row1, row2, row3})
	regexHello := regexp.MustCompile(`hello`)
	// Advance to row3
	_, _ = it.Next() // row1
	_, _ = it.Next() // row2
	row, ok = it.NextUtilRegexIsMatched(regexHello)
	require.True(t, ok)
	require.Equal(t, row3, row)
}
