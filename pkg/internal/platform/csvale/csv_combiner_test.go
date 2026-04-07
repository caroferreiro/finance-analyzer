package csvale

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCombineCSVMatrices(t *testing.T) {
	tests := []struct {
		name      string
		matrices  [][][]string
		want      [][]string
		wantError bool
	}{
		{
			name: "multiple_matrices_with_headers",
			matrices: [][][]string{
				{
					{"Header1", "Header2", "Header3"},
					{"Row1Col1", "Row1Col2", "Row1Col3"},
					{"Row2Col1", "Row2Col2", "Row2Col3"},
				},
				{
					{"Header1", "Header2", "Header3"},
					{"Row3Col1", "Row3Col2", "Row3Col3"},
				},
				{
					{"Header1", "Header2", "Header3"},
					{"Row4Col1", "Row4Col2", "Row4Col3"},
					{"Row5Col1", "Row5Col2", "Row5Col3"},
				},
			},
			want: [][]string{
				{"Header1", "Header2", "Header3"},
				{"Row1Col1", "Row1Col2", "Row1Col3"},
				{"Row2Col1", "Row2Col2", "Row2Col3"},
				{"Row3Col1", "Row3Col2", "Row3Col3"},
				{"Row4Col1", "Row4Col2", "Row4Col3"},
				{"Row5Col1", "Row5Col2", "Row5Col3"},
			},
			wantError: false,
		},
		{
			name: "single_matrix",
			matrices: [][][]string{
				{
					{"Header1", "Header2"},
					{"Row1Col1", "Row1Col2"},
				},
			},
			want: [][]string{
				{"Header1", "Header2"},
				{"Row1Col1", "Row1Col2"},
			},
			wantError: false,
		},
		{
			name:      "empty_matrices_slice",
			matrices:  [][][]string{},
			want:      nil,
			wantError: false,
		},
		{
			name: "matrix_with_empty_rows",
			matrices: [][][]string{
				{
					{"Header1", "Header2"},
					{"Row1Col1", "Row1Col2"},
				},
				{
					{"Header1", "Header2"},
				},
			},
			want: [][]string{
				{"Header1", "Header2"},
				{"Row1Col1", "Row1Col2"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			got := CombineCSVMatrices(tt.matrices)

			// Then
			require.Equal(t, tt.want, got)
		})
	}
}
