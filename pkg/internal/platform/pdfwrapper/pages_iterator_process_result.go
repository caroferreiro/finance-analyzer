package pdfwrapper

// ProcessResult defines the result of processing a word.
type ProcessResult int

const (
	// Continue indicates that processing should continue normally.
	Continue ProcessResult = iota
	// StopIteration indicates that processing should stop (to return an error or to continue iterating in the next function)
	StopIteration
)
