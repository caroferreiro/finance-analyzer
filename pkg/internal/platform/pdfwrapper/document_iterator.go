package pdfwrapper

import (
	"fmt"
)

func NewRealDocumentIterator(doc Document, startingRowPosition int) *RealDocumentIterator {
	return &RealDocumentIterator{
		document:            doc,
		startingRowPosition: startingRowPosition,
		currPage:            0,
		currRow:             0,
	}
}

// RealDocumentIterator is a helper to iterate over the pages of a PDF document. It can be used when you have a function
// that depends on the last position of a function called before, so you can continue processing the pages from that
// position.
// The "Real" prefix is because couldn't think of a better name to distinguish it from the interface and fake.
type RealDocumentIterator struct {
	document Document
	// startingRowPosition can be used to skip the first rows of each page (rows come in descending order)
	startingRowPosition int
	currPage            int
	currRow             int
	currText            int
}

// ContinueProcessingPages continues processing the pages of the document from the last position. It can be used to
// share the iterator between different functions that need to process the pages in order. Where one function left off,
// the next one can continue.
// TODO: reduce code duplication with FoldPages
// Deprecated: use nextText instead to decouple the logic of processing the words from the iterator
func (pi *RealDocumentIterator) ContinueProcessingPages(
	// processWord receives the row word and returns
	// 1. true if it should break from the current page
	// 2. true if it should break from the whole document
	// 3. an error if there was an error processing the word
	processWord func(string) (ProcessResult, error),
) error {
	for iPage := pi.currPage; iPage < len(pi.document.Pages); iPage++ {
		pi.currPage = iPage
		page := pi.document.Pages[iPage]
		pageRows := page.Rows
		for iRow := pi.currRow; iRow < len(pageRows); iRow++ {
			pi.currRow = iRow
			row := pageRows[iRow]
			// Skip this row if it's before the starting row (rows come in descending order)
			if row.Position <= pi.startingRowPosition {
				for iContent := 0; iContent < len(row.Texts); iContent++ {
					text := row.Texts[iContent]
					processResult, err := processWord(text)
					if err != nil {
						return fmt.Errorf("error processing word: %w", err)
					}

					// We have a sequence of ifs instead of a switch because we want to break as soon as we find a break
					if processResult == Continue {
						continue
					}

					if processResult == StopIteration {
						// break from the whole document
						return nil
					}
				}
			}
		}

		// We finished the current page so reset the current row
		pi.currRow = 0
	}
	return nil
}

// NextText returns the next text element. It traverses:
//  1. The next text in the same row,
//  2. The first text in the next row,
//  3. The first text in the first row of the next page.
func (pi *RealDocumentIterator) NextText() (string, bool) {
	// Loop until we exhaust all pages
	for pi.currPage < len(pi.document.Pages) {
		page := pi.document.Pages[pi.currPage]
		// Loop until we exhaust all rows in the current page
		for pi.currRow < len(page.Rows) {
			row := page.Rows[pi.currRow]
			// Process only rows with Position less or equal to startingRowPosition
			if row.Position > pi.startingRowPosition {
				// Skip this row.
				pi.currRow++
				pi.currText = 0
				continue
			}
			// If there is any text left in this row, return it.
			if pi.currText < len(row.Texts) {
				text := row.Texts[pi.currText]
				pi.currText++
				return text, true
			}
			// No more text in this row, so move to the next row.
			pi.currRow++
			pi.currText = 0
		}
		// Finished current page: move to the next page.
		pi.currPage++
		pi.currRow = 0
		pi.currText = 0
	}
	// No more text found in the document.
	return "", false
}
