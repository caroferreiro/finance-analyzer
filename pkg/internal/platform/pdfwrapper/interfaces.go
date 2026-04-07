package pdfwrapper

type DocumentIterator interface {
	NextText() (string, bool)
}
