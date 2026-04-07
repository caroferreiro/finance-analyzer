package pdftable

import (
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
)

// TableIteratorFactory is responsible for creating table iterators with specific configurations.
type TableIteratorFactory struct {
	rowFactory *RowFactory
}

func NewTableIteratorFactory(rowFactory *RowFactory) *TableIteratorFactory {
	return &TableIteratorFactory{
		rowFactory: rowFactory,
	}
}

// CreateIterator creates a new RealTableIterator instance using the factory's configuration
// and the provided document iterator. It pre-parses all rows from the DocumentIterator during
// construction for efficient iteration.
func (f *TableIteratorFactory) CreateIterator(docIterator pdfwrapper.DocumentIterator) *RealTableIterator {
	// Pre-parse all rows
	var rows []Row
	for text, ok := docIterator.NextText(); ok; text, ok = docIterator.NextText() {
		row := f.rowFactory.CreateRow(text)
		rows = append(rows, row)
	}

	return NewRealTableIterator(rows)
}
