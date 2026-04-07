package pdftable

import (
	"regexp"
)

// NewFakeTableIterator creates a new FakeTableIterator with the given rows.
func NewFakeTableIterator(rows []Row) *FakeTableIterator {
	return &FakeTableIterator{
		rows:      rows,
		currIndex: 0,
	}
}

// FakeTableIterator is a test-friendly implementation of a table row iterator that allows preloading
// with a slice of Rows to be returned in sequence.
type FakeTableIterator struct {
	rows      []Row
	currIndex int
}

// Next returns the next Row in the sequence and a boolean indicating if there are more rows to return.
func (f *FakeTableIterator) Next() (Row, bool) {
	if f.currIndex >= len(f.rows) {
		return Row{}, false
	}
	row := f.rows[f.currIndex]
	f.currIndex++
	return row, true
}

// NextUtilRegexIsMatched returns the next Row whose RawText matches the given regex.
func (f *FakeTableIterator) NextUtilRegexIsMatched(regex *regexp.Regexp) (Row, bool) {
	for f.currIndex < len(f.rows) {
		row := f.rows[f.currIndex]
		f.currIndex++
		if regex.MatchString(row.RawText) {
			return row, true
		}
	}
	return Row{}, false
}
