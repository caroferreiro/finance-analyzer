package pdfwrapper

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestFoldPages(t *testing.T) {
	type testCase struct {
		name        string
		pages       []Page
		processWord func(string) (int, error)
		joinFn      func(int, int) (int, error)
		expected    int
		// expectedErr is a substring that should appear in the error message, if any.
		expectedErr error
	}

	tests := []testCase{
		{
			name:        "Empty Pages",
			pages:       []Page{},
			processWord: nil,
			joinFn:      nil,
			expected:    0,
			expectedErr: ErrNoPages,
		},
		{
			name: "Single Word",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						{Position: 0, Texts: []string{"5"}},
					},
				},
			},
			processWord: func(s string) (int, error) {
				return strconv.Atoi(s)
			},
			joinFn: func(a, b int) (int, error) {
				return a + b, nil
			},
			expected:    5,
			expectedErr: nil,
		},
		{
			name: "Multiple Words",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						{Position: 0, Texts: []string{"1", "2"}},
						{Position: 1, Texts: []string{"3"}},
					},
				},
			},
			processWord: func(s string) (int, error) {
				return strconv.Atoi(s)
			},
			joinFn: func(a, b int) (int, error) {
				return a + b, nil
			},
			expected:    6,
			expectedErr: nil,
		},
		{
			name: "ProcessWord Error",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						{Position: 0, Texts: []string{"hello", "error", "world"}},
					},
				},
			},
			processWord: func(s string) (int, error) {
				if s == "error" {
					return 0, errors.New("process error")
				}
				return 1, nil
			},
			joinFn: func(a, b int) (int, error) {
				return a + b, nil
			},
			expected:    0,
			expectedErr: fmt.Errorf("error processing word: %w", errors.New("process error")),
		},
		{
			name: "JoinFn Error",
			pages: []Page{
				{
					Index: 0,
					Rows: []Row{
						{Position: 0, Texts: []string{"a", "errorjoin"}},
					},
				},
			},
			processWord: func(s string) (int, error) {
				// Return the length of the word.
				return len(s), nil
			},
			joinFn: func(a, b int) (int, error) {
				// Return error when b equals 9.
				if b == 9 {
					return 0, errors.New("join error")
				}
				return a + b, nil
			},
			expected:    0,
			expectedErr: fmt.Errorf("error joining results: %w", errors.New("join error")),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// When
			actualResult, actualErr := FoldPages(tc.pages, tc.processWord, tc.joinFn)

			// Then
			require.Equal(t, tc.expected, actualResult)
			require.Equal(t, tc.expectedErr, actualErr)
		})
	}
}
