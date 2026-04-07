package pdftable

import (
	"regexp"
)

func NewRealTableIterator(rows []Row) *RealTableIterator {
	return &RealTableIterator{
		rows:      rows,
		currIndex: 0,
	}
}

// RealTableIterator is used to iterate the rows of a PDF table. The idea is to share this iterator between different
// functions, so it can continue where the old one finished.
// The "Real" prefix is because couldn't think of a better name to distinguish it from the interface and fake.
type RealTableIterator struct {
	// rows holds the pre-parsed rows as values for consistency with FakeTableIterator
	rows []Row
	// currIndex tracks the current position in the rows slice
	currIndex int
}

// Next implements the TableIterator interface. It returns the next Row in the sequence
// and a boolean indicating if there are more rows to return.
func (it *RealTableIterator) Next() (Row, bool) {
	if it.currIndex >= len(it.rows) {
		return Row{}, false
	}
	row := it.rows[it.currIndex]
	it.currIndex++
	return row, true
}

// NextUtilRegexIsMatched implements the TableIterator interface. It iterates through the rows
// until a row matches the regex and returns it.
func (it *RealTableIterator) NextUtilRegexIsMatched(regex *regexp.Regexp) (Row, bool) {
	for it.currIndex < len(it.rows) {
		row := it.rows[it.currIndex]
		it.currIndex++
		if regex.MatchString(row.RawText) {
			return row, true
		}
	}
	return Row{}, false
}
