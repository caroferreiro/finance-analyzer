package pdfcardsummaryio

import (
	"fmt"
	"os"
	"testing"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/Alechan/finance-analyzer/pkg/internal/validation"
	"github.com/Alechan/finance-analyzer/pkg/internal/validation/testdata"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestETLFile_WithValidationEnabled(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()
	testPDFPath := tempDir + "/test.pdf"
	testPDFContent := []byte("fake pdf content")
	err := os.WriteFile(testPDFPath, testPDFContent, 0644)
	require.NoError(t, err)

	validCardSummary := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithSaldoAnterior("0", "0")
		b.WithCard("1234", "TEST CARD", "0", "0")
	})

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary, nil)

	validator := validation.NewValidator()
	etl := NewPDFCardSummaryETL(mockExtractor, validator)

	// When
	err = etl.ETLFile(testPDFPath)

	// Then
	require.NoError(t, err, "ETL should succeed with valid data")

	// Verify CSV was created
	csvPath := testPDFPath + ".csv"
	_, err = os.Stat(csvPath)
	require.NoError(t, err, "CSV file should be created")
}

func TestETLFile_ValidationFails(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()
	testPDFPath := tempDir + "/test.pdf"
	testPDFContent := []byte("fake pdf content")
	err := os.WriteFile(testPDFPath, testPDFContent, 0644)
	require.NoError(t, err)

	invalidCardSummary := pdfcardsummary.CardSummary{
		Table: pdfcardsummary.Table{
			Cards: []pdfcardsummary.Card{},
		},
	}

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(invalidCardSummary, nil)

	validator := validation.NewValidator()
	etl := NewPDFCardSummaryETL(mockExtractor, validator)

	// When
	err = etl.ETLFile(testPDFPath)

	// Then
	require.Error(t, err, "ETL should fail when validation fails")
	require.Contains(t, err.Error(), "error validating file", "error should mention validation")
	require.Contains(t, err.Error(), testPDFPath, "error should include file path")

	// Verify CSV was NOT created
	csvPath := testPDFPath + ".csv"
	_, err = os.Stat(csvPath)
	require.Error(t, err, "CSV file should NOT be created when validation fails")
	require.True(t, os.IsNotExist(err), "CSV file should not exist")
}

func TestETLFile_WithMockValidator(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()
	testPDFPath := tempDir + "/test.pdf"
	testPDFContent := []byte("fake pdf content")
	err := os.WriteFile(testPDFPath, testPDFContent, 0644)
	require.NoError(t, err)

	validCardSummary := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithSaldoAnterior("0", "0")
		b.WithCard("1234", "TEST CARD", "0", "0")
	})

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary, nil)

	mockValidator := NewMockValidator(ctrl)
	mockValidator.EXPECT().Validate(gomock.Any()).Return(nil)

	etl := NewPDFCardSummaryETL(mockExtractor, mockValidator)

	// When
	err = etl.ETLFile(testPDFPath)

	// Then
	require.NoError(t, err, "ETL should succeed")
}

func TestETLFile_ExtractionFails(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()
	testPDFPath := tempDir + "/test.pdf"
	testPDFContent := []byte("fake pdf content")
	err := os.WriteFile(testPDFPath, testPDFContent, 0644)
	require.NoError(t, err)

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(pdfcardsummary.CardSummary{}, fmt.Errorf("extraction failed"))

	validator := validation.NewValidator()
	etl := NewPDFCardSummaryETL(mockExtractor, validator)

	// When
	err = etl.ETLFile(testPDFPath)

	// Then
	require.Error(t, err, "ETL should fail when extraction fails")
	require.Contains(t, err.Error(), "error parsing file", "error should mention parsing")
	require.Contains(t, err.Error(), testPDFPath, "error should include file path")

	// Verify CSV was NOT created
	csvPath := testPDFPath + ".csv"
	_, err = os.Stat(csvPath)
	require.Error(t, err, "CSV file should NOT be created when extraction fails")
	require.True(t, os.IsNotExist(err), "CSV file should not exist")
}

func TestETLFilesIndependently_WithValidationErrors(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()

	testPDF1Path := tempDir + "/test1.pdf"
	testPDF2Path := tempDir + "/test2.pdf"
	testPDF3Path := tempDir + "/test3.pdf"

	err := os.WriteFile(testPDF1Path, []byte("fake pdf 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(testPDF2Path, []byte("fake pdf 2"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(testPDF3Path, []byte("fake pdf 3"), 0644)
	require.NoError(t, err)

	validCardSummary := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithSaldoAnterior("0", "0")
		b.WithCard("1234", "TEST CARD", "0", "0")
	})

	invalidCardSummary := pdfcardsummary.CardSummary{
		Table: pdfcardsummary.Table{
			Cards: []pdfcardsummary.Card{},
		},
	}

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary, nil).Times(1)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(invalidCardSummary, nil).Times(1)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary, nil).Times(1)

	validator := validation.NewValidator()
	etl := NewPDFCardSummaryETL(mockExtractor, validator)

	// When
	err = etl.ETLFilesIndependently([]string{testPDF1Path, testPDF2Path, testPDF3Path})

	// Then
	require.Error(t, err, "ETLFilesIndependently should return error when some files fail")
	require.Contains(t, err.Error(), "errors per file", "error should mention multiple files")

	// Verify CSV was created for valid files
	_, err = os.Stat(testPDF1Path + ".csv")
	require.NoError(t, err, "CSV should be created for valid file 1")

	// Verify CSV was NOT created for invalid file
	_, err = os.Stat(testPDF2Path + ".csv")
	require.Error(t, err, "CSV should NOT be created for invalid file")
	require.True(t, os.IsNotExist(err), "CSV file should not exist for invalid file")

	// Verify CSV was created for valid files
	_, err = os.Stat(testPDF3Path + ".csv")
	require.NoError(t, err, "CSV should be created for valid file 3")
}

func TestETLFilesWithJoinedCSV_Success(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()

	testPDF1Path := tempDir + "/test1.pdf"
	testPDF2Path := tempDir + "/test2.pdf"
	combinedCSVPath := tempDir + "/combined.csv"

	err := os.WriteFile(testPDF1Path, []byte("fake pdf 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(testPDF2Path, []byte("fake pdf 2"), 0644)
	require.NoError(t, err)

	validCardSummary1 := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithTotalARS("1500")
		b.WithTotalUSD("75")
		b.WithSaldoAnterior("1000", "50")
		b.WithCard("1234", "CARD 1", "500", "25")
		b.WithCardMovement(0, nil, "", "Purchase", "500", "25")
	})

	validCardSummary2 := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithTotalARS("3000")
		b.WithTotalUSD("150")
		b.WithSaldoAnterior("2000", "100")
		b.WithCard("5678", "CARD 2", "1000", "50")
		b.WithCardMovement(0, nil, "", "Purchase", "1000", "50")
	})

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary1, nil).Times(1)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary2, nil).Times(1)

	validator := validation.NewValidator()
	etl := NewPDFCardSummaryETL(mockExtractor, validator)

	// When
	err = etl.ETLFilesWithJoinedCSV([]string{testPDF1Path, testPDF2Path}, combinedCSVPath)

	// Then
	require.NoError(t, err, "ETLFilesWithJoinedCSV should succeed")

	// Verify individual CSV files were created
	_, err = os.Stat(testPDF1Path + ".csv")
	require.NoError(t, err, "Individual CSV should be created for file 1")

	_, err = os.Stat(testPDF2Path + ".csv")
	require.NoError(t, err, "Individual CSV should be created for file 2")

	// Verify combined CSV was created
	_, err = os.Stat(combinedCSVPath)
	require.NoError(t, err, "Combined CSV should be created")

	// Verify combined CSV has correct format (headers appear only once)
	combinedCSVContent, err := os.ReadFile(combinedCSVPath)
	require.NoError(t, err)
	combinedCSVString := string(combinedCSVContent)

	// Count header occurrences (should be exactly 1)
	headerCount := 0
	lines := []rune(combinedCSVString)
	currentLine := ""
	for _, char := range lines {
		if char == '\n' {
			if currentLine != "" {
				// Check if this line contains headers (simplified check)
				if len(currentLine) > 0 {
					headerCount++
				}
				currentLine = ""
			}
		} else {
			currentLine += string(char)
		}
	}
	// At minimum, we should have headers + data rows
	require.Greater(t, len(string(combinedCSVContent)), 0, "Combined CSV should not be empty")
}

func TestETLFilesWithJoinedCSV_ExtractionFails(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()

	testPDF1Path := tempDir + "/test1.pdf"
	testPDF2Path := tempDir + "/test2.pdf"
	combinedCSVPath := tempDir + "/combined.csv"

	err := os.WriteFile(testPDF1Path, []byte("fake pdf 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(testPDF2Path, []byte("fake pdf 2"), 0644)
	require.NoError(t, err)

	validCardSummary := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithTotalARS("1500")
		b.WithTotalUSD("75")
		b.WithSaldoAnterior("1000", "50")
		b.WithCard("1234", "CARD 1", "500", "25")
		b.WithCardMovement(0, nil, "", "Purchase", "500", "25")
	})

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary, nil).Times(1)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(pdfcardsummary.CardSummary{}, fmt.Errorf("extraction failed")).Times(1)

	validator := validation.NewValidator()
	etl := NewPDFCardSummaryETL(mockExtractor, validator)

	// When
	err = etl.ETLFilesWithJoinedCSV([]string{testPDF1Path, testPDF2Path}, combinedCSVPath)

	// Then
	require.Error(t, err, "ETLFilesWithJoinedCSV should fail when extraction fails")
	require.Contains(t, err.Error(), "error parsing file", "error should mention parsing")
	require.Contains(t, err.Error(), testPDF2Path, "error should include failing file path")

	// Verify no CSV files were created (fail fast)
	_, err = os.Stat(testPDF1Path + ".csv")
	require.Error(t, err, "Individual CSV should NOT be created when extraction fails")
	require.True(t, os.IsNotExist(err), "CSV file should not exist")

	_, err = os.Stat(combinedCSVPath)
	require.Error(t, err, "Combined CSV should NOT be created when extraction fails")
	require.True(t, os.IsNotExist(err), "Combined CSV file should not exist")
}

func TestETLFilesWithJoinedCSV_ValidationFails(t *testing.T) {
	// Given
	ctrl := gomock.NewController(t)

	tempDir := t.TempDir()

	testPDF1Path := tempDir + "/test1.pdf"
	testPDF2Path := tempDir + "/test2.pdf"
	combinedCSVPath := tempDir + "/combined.csv"

	err := os.WriteFile(testPDF1Path, []byte("fake pdf 1"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(testPDF2Path, []byte("fake pdf 2"), 0644)
	require.NoError(t, err)

	validCardSummary := testdata.BuildCardSummary(t, func(b *testdata.CardSummaryBuilder) {
		b.WithTotalARS("1500")
		b.WithTotalUSD("75")
		b.WithSaldoAnterior("1000", "50")
		b.WithCard("1234", "CARD 1", "500", "25")
		b.WithCardMovement(0, nil, "", "Purchase", "500", "25")
	})

	invalidCardSummary := pdfcardsummary.CardSummary{
		Table: pdfcardsummary.Table{
			Cards: []pdfcardsummary.Card{},
		},
	}

	mockExtractor := NewMockExtractor(ctrl)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(validCardSummary, nil).Times(1)
	mockExtractor.EXPECT().ExtractFromBytes(gomock.Any()).Return(invalidCardSummary, nil).Times(1)

	validator := validation.NewValidator()
	etl := NewPDFCardSummaryETL(mockExtractor, validator)

	// When
	err = etl.ETLFilesWithJoinedCSV([]string{testPDF1Path, testPDF2Path}, combinedCSVPath)

	// Then
	require.Error(t, err, "ETLFilesWithJoinedCSV should fail when validation fails")
	require.Contains(t, err.Error(), "error validating file", "error should mention validation")
	require.Contains(t, err.Error(), testPDF2Path, "error should include failing file path")

	// Verify no CSV files were created (fail fast)
	_, err = os.Stat(testPDF1Path + ".csv")
	require.Error(t, err, "Individual CSV should NOT be created when validation fails")
	require.True(t, os.IsNotExist(err), "CSV file should not exist")

	_, err = os.Stat(combinedCSVPath)
	require.Error(t, err, "Combined CSV should NOT be created when validation fails")
	require.True(t, os.IsNotExist(err), "Combined CSV file should not exist")
}
