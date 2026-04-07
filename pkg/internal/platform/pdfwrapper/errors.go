package pdfwrapper

import (
	"errors"
)

var (
	ErrNilOrEmptyRawBytes   = errors.New("nil or empty raw bytes")
	ErrCreatingVendorReader = errors.New("couldn't create vendor pdf reader")
	ErrNoPages              = errors.New("no pages provided")
	ErrPatternNotFound      = errors.New("reached the end of the file without finding the expected pattern")
)
