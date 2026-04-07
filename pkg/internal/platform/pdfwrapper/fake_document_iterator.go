package pdfwrapper

// NewFakeDocumentIterator creates a new FakeDocumentIterator with the given texts.
func NewFakeDocumentIterator(texts []string) *FakeDocumentIterator {
	return &FakeDocumentIterator{
		texts:     texts,
		currIndex: 0,
	}
}

// FakeDocumentIterator is a test-friendly implementation of DocumentIterator that allows preloading
// with a slice of strings to be returned in sequence.
type FakeDocumentIterator struct {
	// texts holds the preloaded texts to be returned in sequence
	texts []string
	// currIndex tracks the current position in the texts slice
	currIndex int
}

// NextText implements the DocumentIterator interface. It returns the next text in the sequence
// and a boolean indicating if there are more texts to return.
func (f *FakeDocumentIterator) NextText() (string, bool) {
	if f.currIndex >= len(f.texts) {
		return "", false
	}
	text := f.texts[f.currIndex]
	f.currIndex++
	return text, true
}
