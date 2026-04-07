package pdftable

import "regexp"

type TableIterator interface {
	// Next returns the next Row in the sequence and a boolean indicating if there are more rows to return.
	Next() (Row, bool)
	// NextUtilRegexIsMatched iterates through the rows until a row matches the regex and returns it
	NextUtilRegexIsMatched(regex *regexp.Regexp) (Row, bool)
}
