package pdfwrapper

func NewDocument(pages []Page) Document {
	return Document{
		Pages: pages,
	}
}

// Document represents the whole PDF document. It contains all the pages.
type Document struct {
	Pages []Page
}
