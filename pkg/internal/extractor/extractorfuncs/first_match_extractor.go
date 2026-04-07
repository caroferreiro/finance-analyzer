package extractorfuncs

import (
	"fmt"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pdfwrapper"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pointersale"
	"regexp"
)

func NewFirstMatchExtractor[T any](
	regex *regexp.Regexp,
	deserializerFn func(rawStr string) (T, error),
) *FirstMatchExtractor[T] {
	return &FirstMatchExtractor[T]{
		regex:          regex,
		deserializerFn: deserializerFn,
	}
}

type FirstMatchExtractor[T any] struct {
	regex          *regexp.Regexp
	deserializerFn func(rawStr string) (T, error)
}

func (de *FirstMatchExtractor[T]) ExtractFirstMatch(pages []pdfwrapper.Page) (T, error) {
	potentialResult, err := pdfwrapper.FoldPages(
		pages,
		// The function that will be called for each text in the pages
		de.lineMatcher,
		// The join function (it will just keep the first non nil match found)
		keepFirstIfNonNil,
	)
	if err != nil {
		return *new(T), fmt.Errorf("error running fold on pages: %w", err)
	}

	if potentialResult == nil {
		return *new(T), fmt.Errorf("error extracting first match: %w", pdfwrapper.ErrPatternNotFound)
	}
	return *potentialResult, nil
}

// lineMatcher is a function that will be called for each text in the pages. It returns a pointer to T because
// we need to see if it was able to extract a value or not.
func (de *FirstMatchExtractor[T]) lineMatcher(text string) (*T, error) {
	matches := de.regex.FindStringSubmatch(text)
	if len(matches) > 1 {
		rawDate := matches[1]
		candidate, err := de.deserializerFn(rawDate)
		if err != nil {
			return nil, fmt.Errorf(
				"error extracting from '%s' the value '%s' and converting to type %T: %w",
				text,
				rawDate,
				candidate,
				err,
			)
		}
		return pointersale.ToPointer(candidate), nil
	}
	return nil, nil
}

func keepFirstIfNonNil[T any](t1, t2 *T) (*T, error) {
	if t1 != nil {
		return t1, nil
	}
	return t2, nil
}
