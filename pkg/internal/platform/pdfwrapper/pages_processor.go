package pdfwrapper

import "fmt"

// ProcessPagesOld is a helper function that processes the rows of the pages of the PDF, calling the processWord function for each word.
// The inner function return values should be:
// 1. true if it should break from the current page
// 2. true if it should break from the whole document
// 3. an error if there was an error processing the word
//
// The outer functions returns
// 1. The error if there was one. If the processing covers all the pages without breaking the loop, it returns an error
// TODO: reduce code duplication with pages iterator
// Deprecated: use FoldPages instead
func ProcessPagesOld(pages []Page, processWord func(string) (bool, bool, error)) error {
	if len(pages) == 0 {
		return ErrNoPages
	}

	for _, page := range pages {
		for _, row := range page.Rows {
			for i := 0; i < len(row.Texts); i++ {
				text := row.Texts[i]
				shouldBreakPage, shouldBreakDoc, err := processWord(text)
				if err != nil {
					return fmt.Errorf("error processing word: %w", err)
				}
				if shouldBreakPage {
					break
				}
				if shouldBreakDoc {
					return nil
				}

			}
		}
	}
	return ErrPatternNotFound
}

// FoldPages is a helper function that processes the rows of the pages of the PDF, calling the processWord function for each word.
// It's called "fold" because it's a common pattern in functional programming to fold a list of elements into a single value.
// https://en.wikipedia.org/wiki/Fold_(higher-order_function)
// TODO: reduce code duplication with pages iterator
func FoldPages[T any](
	pages []Page,
	// processWord should be provided by the caller and will be called for each word in the PDF. It should return
	// 1. The value of the type T will be returned by FoldPages
	// 2. An error if there was an error processing the word
	processWord func(string) (T, error),
	// joinFn should be provided by the caller and will be called to join the values of the type T returned by processWord
	// It's the responsibility of the caller to use this function to handle the different results of processWord
	joinFn func(T, T) (T, error),
) (T, error) {
	if len(pages) == 0 {
		return *new(T), ErrNoPages
	}

	// Keep track of the accumulated result. It will be the result of sequentially joining the "individual results" of
	// processing each word
	var accumResult T

	for _, page := range pages {
		for _, row := range page.Rows {
			for _, text := range row.Texts {
				individualResult, err := processWord(text)
				if err != nil {
					return *new(T), fmt.Errorf("error processing word: %w", err)
				}

				accumResult, err = joinFn(accumResult, individualResult)
				if err != nil {
					return *new(T), fmt.Errorf("error joining results: %w", err)
				}
			}
		}
	}
	return accumResult, nil
}
