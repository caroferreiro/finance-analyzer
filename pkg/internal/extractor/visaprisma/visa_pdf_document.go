package visaprisma

import (
	"fmt"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/pointersale"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/stringsale"
	"github.com/Alechan/finance-analyzer/pkg/internal/platform/timeale"
	"github.com/Alechan/pdf"
	"github.com/shopspring/decimal"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	cierreActualRegex              = regexp.MustCompile(`CIERRE ACTUAL: (\d{2} \w{3} \d{2})`)
	totalConsumosTarjetaRegex      = regexp.MustCompile(`^\s*(?:Tarjeta\s+(\d+)\s+)?Total Consumos de\s+([A-Z0-9\s]+)\s+([\d\.,]+)\s+([\d\.,]+-?)`)
	tableHeadersRegex              = regexp.MustCompile(`^\s*FECHA\s+COMPROBANTE\s+DETALLE\sDE\sTRANSACCION\s+PESOS\s+DOLARES\s*$`)
	saldoAnteriorRegex             = regexp.MustCompile(`\s*SALDO\sANTERIOR\s+([\d.,]+)\s+([\d.,]+-?)`)
	detailWithInstallmentsRegex    = regexp.MustCompile(`^(.+)\s+Cuota\s+(\d{2})/(\d{2})$`)
	cardMovementRegex              = regexp.MustCompile(`\d{2}\.\d{2}\.\d{2}\s+([A-Za-z0-9*]+)?\s+([A-Za-z\s%.\/,0-9\-$]+)\s+([0-9,.]+)`)
	isNumberRegex                  = regexp.MustCompile(`\d`)
	isNumberOrSignRegex            = regexp.MustCompile(`\d|-`)
	isFinalRowWithSaldoActualRegex = regexp.MustCompile(`SALDO ACTUAL\s+\$\s+(\d{1,3}(?:\.\d{3})*,\d{2})(?:\s+U\$S\s+(\d{1,3}(?:\.\d{3})*,\d{2}-?))?`)
)

const movsGroupEndCharacter = "_"

func NewDocumentFromFilePath(filePath string) (Document, error) {
	// Read the file into a byte slice
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Document{}, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return NewDocumentFromBytes(data)
}

func NewDocumentFromBytes(rawBytes []byte) (Document, error) {
	readerAt := strings.NewReader(string(rawBytes))
	return NewDocumentFromReaderAt(readerAt, int64(len(rawBytes)))
}

func NewDocumentFromReaderAt(readerAt io.ReaderAt, size int64) (Document, error) {
	allRows, err := readPDFRows(readerAt, size)
	if err != nil {
		return Document{}, fmt.Errorf("error reading pdf rows: %w", err)
	}

	return NewDocumentFromPDFRows(allRows)
}

func readPDFRows(readerAt io.ReaderAt, size int64) ([]*pdf.Row, error) {
	if readerAt == nil {
		return nil, fmt.Errorf("readerAt is nil")
	}

	r, err := pdf.NewReader(readerAt, size)
	if err != nil {
		return nil, fmt.Errorf("couldn't create pdf reader: %w", err)
	}

	allRows, err := extractAllRows(r)
	if err != nil {
		return nil, fmt.Errorf("error extracting all rows: %w", err)
	}
	return allRows, nil
}

func NewDocumentFromPDFRows(rows []*pdf.Row) (Document, error) {

	// We're going to read the PDF rows in order and extract the information piece by piece
	i := 0
	maxRowIdx := len(rows) - 1

	// Find the row with the closing date
	closingDate, err := extractClosingDate(rows, i, maxRowIdx)
	if err != nil {
		return Document{}, fmt.Errorf("error finding the closing date: %w", err)
	}

	// Find the row with the expiration date
	i, expDate, err := extractExpirationDate(rows, i, maxRowIdx)
	if err != nil {
		return Document{}, fmt.Errorf("error finding the expiration date: %w", err)
	}

	// Find the start of the table
	i, j, err := findTableHeadersPosition(rows, i, maxRowIdx)
	if err != nil {
		return Document{}, fmt.Errorf("error finding table headers: %w", err)
	}

	// The table doesn't have separators between columns, so we need to find the position of each column
	tablePositions, err := findTableColumnPositions(rows, i, maxRowIdx, j)
	if err != nil {
		return Document{}, fmt.Errorf("error finding table column positions: %w", err)
	}

	// Parse the table with the movements per card and the total at the end
	// The first rows are movements with the payment information of the previous card window
	i, j, pastPaymentMovements, err := extractPastPaymentMovements(rows, i, maxRowIdx, j, tablePositions)
	if err != nil {
		return Document{}, fmt.Errorf("error extracting past payment movements: %w", err)
	}

	cards, cardsOrphanMovements, i, j, err := extractPDFCards(rows, i, maxRowIdx, j, tablePositions)
	if err != nil {
		return Document{}, fmt.Errorf("error extracting pdf cards: %w", err)
	}

	totalARS, totalUSD, err := extractTotalAmountsOfWholeDocument(rows, i, j)
	if err != nil {
		return Document{}, fmt.Errorf("error extracting total amounts: %w", err)
	}

	// TODO: add validations before returning the document
	doc := Document{
		TotalARS:             totalARS,
		TotalUSD:             totalUSD,
		CloseDate:            closingDate,
		ExpirationDate:       expDate,
		PastPaymentMovements: pastPaymentMovements,
		Cards:                cards,
		TaxesMovements:       cardsOrphanMovements,
		TablePositions:       tablePositions,
	}

	return doc, nil
}

func extractTotalAmountsOfWholeDocument(rows []*pdf.Row, i int, j int) (decimal.Decimal, decimal.Decimal, error) {
	text := extractRowWord(rows, i, j)
	matchesSaldoActual := isFinalRowWithSaldoActualRegex.FindStringSubmatch(text)
	if len(matchesSaldoActual) < 3 {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("unexpected end of file. Couldn't find the total amounts")
	}

	totalARS, err := PDFAmountToDecimal(matchesSaldoActual[1])
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("error converting doc total ARS amount %s to decimal: %w", matchesSaldoActual[1], err)
	}

	totalUSD, err := PDFAmountToDecimal(matchesSaldoActual[2])
	if err != nil {
		return decimal.Decimal{}, decimal.Decimal{}, fmt.Errorf("error converting doc total USD amount %s to decimal: %w", matchesSaldoActual[2], err)
	}
	return totalARS, totalUSD, nil
}

func extractClosingDate(rows []*pdf.Row, i int, maxRowIdx int) (time.Time, error) {
	var closingDate time.Time
	i, _, err := processRows(
		rows,
		i,
		maxRowIdx,
		0,
		func(text string) (bool, error) {
			matchesCierreActual := cierreActualRegex.FindStringSubmatch(text)
			if len(matchesCierreActual) > 1 {
				rawClosingDate := matchesCierreActual[1]
				closingDateAsTime, err := timeale.CardSummarySpanishMonthDateToTime(rawClosingDate)
				if err != nil {
					return true, fmt.Errorf("error converting closing date %s to time: %w", rawClosingDate, err)
				}

				closingDate = closingDateAsTime

				return true, nil
			}

			return false, nil
		},
	)
	if err != nil {
		return time.Time{}, fmt.Errorf("error extracting closing date: %w", err)
	}
	return closingDate, nil
}

func extractExpirationDate(rows []*pdf.Row, i int, maxRowIdx int) (int, time.Time, error) {
	for i < maxRowIdx {
		if rows[i].Content[0].S == "VENCIMIENTO" {
			rawExpirationDate := rows[i+1].Content[0].S
			expirationDateAsTime, err := timeale.CardSummarySpanishMonthDateToTime(rawExpirationDate)
			if err != nil {
				return 0, time.Time{}, fmt.Errorf("error converting expiration date %s to time: %w", rawExpirationDate, err)
			}

			return i, expirationDateAsTime, nil
		}
		i++
	}
	return 0, time.Time{}, fmt.Errorf("unexpected end of file after looking for the expiration date")
}

func findTableColumnPositions(rows []*pdf.Row, i int, maxRowIdx int, j int) (PDFTablePositions, error) {
	// Even though he rows of the table don't have a separator between columns, they have a fixed width. If a cell is
	// empty, it will be filled with spaces (unless it's the last cell of the row).
	// We should be able to find several "columns representatives" that we can use to find the position of each column.

	// For the amounts, we know that the "total amount per card" includes both, even if 0
	lastArsAmountPosition, lastUsdAmountPosition, err := findAmountsPositions(rows, i, maxRowIdx, j)
	if err != nil {
		return PDFTablePositions{}, fmt.Errorf("error finding amounts positions: %w", err)
	}

	// For the detail, we know that the "previous balance row" doesn't have a date or receipt number, so we can use it to find
	// the start of the detail
	detailStart, err := findDetailStartPosition(rows, i, maxRowIdx, j, lastArsAmountPosition)

	// The start of the date should always be 0 because it's the first column (and we trim left spaces)
	// The date always has the same length, so we can find the end of the date by adding the length of the date to the start
	dateStart := 0
	dateEnd := dateStart + 7

	// It's hard to find the start and end of the receipt number, but we already know end and start of the columns around it,
	// so we can calculate it (the details start and end will probably point to whitespaces but we can ignore them)
	receiptStart := dateEnd + 1
	receiptEnd := detailStart - 1

	return PDFTablePositions{
		OriginalDateStart: dateStart,
		OriginalDateEnd:   dateEnd,
		ReceiptStart:      receiptStart,
		ReceiptEnd:        receiptEnd,
		DetailStart:       detailStart,
		ARSAmountStart:    lastArsAmountPosition - 15,
		ARSAmountEnd:      lastArsAmountPosition,
		USDAmountStart:    lastArsAmountPosition + 1,
		USDAmountEnd:      lastUsdAmountPosition,
	}, nil
}

// processRows is a helper function that processes the rows of the PDF, calling the processWord function for each word.
// The inner function should return true if the processing should stop, and an error if there was an error processing the word.
// The outer functions returns
// 1. The last i
// 2. The last j
// 3. The error if there was one
// DEPRECATED: use the generic one from pdfwrapper
func processRows(
	rows []*pdf.Row,
	i int,
	maxRowIdx int,
	j int,
	processWord func(string) (bool, error),
) (int, int, error) {
	for i < maxRowIdx {
		r := rows[i]
		for j < len(r.Content) {
			text := extractRowWord(rows, i, j)
			shouldBreak, err := processWord(text)
			if err != nil {
				return i, j, fmt.Errorf("error processing word: %w", err)
			}
			if shouldBreak {
				return i, j, nil
			}
			j++
		}
		j = 0
		i++
	}
	return i, j, fmt.Errorf("reached the end of the file without finding the expected pattern")
}

func extractRowWord(rows []*pdf.Row, i int, j int) string {
	r := rows[i]
	word := r.Content[j]
	text := strings.TrimSpace(word.S)
	return text
}

func findAmountEndIndex(text string, potentialIndex int) (int, error) {
	shortLastArsAmountChar := text[potentialIndex]
	if !isNumberOrSignRegex.MatchString(string(shortLastArsAmountChar)) {
		return 0, fmt.Errorf("unexpected character at the end of the ARS amount: %s", string(shortLastArsAmountChar))
	}

	// If the last character is a number, then it's not the "theoretical" end of the amount because it could
	// be followed by a sign
	if isNumberRegex.MatchString(string(shortLastArsAmountChar)) {
		potentialIndex++
	}

	return potentialIndex, nil
}

func findDetailStartPosition(rows []*pdf.Row, i int, maxRowIdx int, j int, standardLastArsAmountPosition int) (int, error) {
	var detailStart int
	_, _, err := processRows(
		rows,
		i,
		maxRowIdx,
		j,
		func(text string) (bool, error) {
			indices := saldoAnteriorRegex.FindStringSubmatchIndex(text)
			if len(indices) > 1 {
				shortLastArsAmountPosition, err := findAmountEndIndex(text, indices[3]-1)
				if err != nil {
					return true, fmt.Errorf("error finding the end of the short ARS amount: %w", err)
				}

				// The detail position is (long end ARS) - (short end ARS)
				detailStart = standardLastArsAmountPosition - shortLastArsAmountPosition
				return true, nil
			}
			return false, nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("error finding detail start position: %w", err)
	}

	return detailStart, nil
}

func extractPDFCards(rows []*pdf.Row, i int, maxRowIdx int, j int, tablePositions PDFTablePositions) ([]PDFCard, []PDFMovement, int, int, error) {
	var pdfCards []PDFCard
	var cardMovs []PDFMovement
	rowFn := func(text string) (bool, error) {
		if cardMovementRegex.MatchString(text) {
			mov, err := extractMovFromText(text, tablePositions)
			if err != nil {
				return true, fmt.Errorf("error extracting card movement: %w", err)
			}
			cardMovs = append(cardMovs, mov)
			return false, nil
		}

		matchesTotalConsumosTarjeta := totalConsumosTarjetaRegex.FindStringSubmatch(text)
		if len(matchesTotalConsumosTarjeta) > 1 {
			cardArsTotal, err := PDFAmountToDecimal(strings.TrimSpace(matchesTotalConsumosTarjeta[3]))
			if err != nil {
				return true, fmt.Errorf("error converting card ARS total %s to decimal: %w", matchesTotalConsumosTarjeta[3], err)
			}
			cardUSDTotal, err := PDFAmountToDecimal(strings.TrimSpace(matchesTotalConsumosTarjeta[4]))
			if err != nil {
				return true, fmt.Errorf("error converting card USD total %s to decimal: %w", matchesTotalConsumosTarjeta[4], err)
			}
			card, err := NewPDFCard(
				strings.TrimSpace(matchesTotalConsumosTarjeta[1]),
				strings.TrimSpace(matchesTotalConsumosTarjeta[2]),
				cardMovs,
				cardArsTotal,
				cardUSDTotal,
			)
			if err != nil {
				return true, fmt.Errorf("error creating new pdf card: %w", err)
			}

			cardMovs = nil
			pdfCards = append(pdfCards, card)

			return false, nil
		}

		if isFinalRowWithSaldoActualRegex.MatchString(text) {
			return true, nil
		}

		return false, nil
	}

	finalI, finalJ, err := processRows(
		rows,
		i,
		maxRowIdx,
		j,
		rowFn,
	)
	if err != nil {
		return nil, nil, 0, 0, fmt.Errorf("error calling processRows: %w", err)
	}

	return pdfCards, cardMovs, finalI, finalJ, nil
}

func findAmountsPositions(rows []*pdf.Row, i int, maxRowIdx int, j int) (int, int, error) {
	var lastArsAmountPosition int
	var lastUsdAmountPosition int
	_, _, err := processRows(
		rows,
		i,
		maxRowIdx,
		j,
		func(text string) (bool, error) {
			indices := totalConsumosTarjetaRegex.FindStringSubmatchIndex(text)
			if len(indices) > 1 {
				innerLastArsAmountPosition, err := findAmountEndIndex(text, indices[7]-1)
				if err != nil {
					return true, fmt.Errorf("error finding the end of the ARS amount: %w", err)
				}
				lastArsAmountPosition = innerLastArsAmountPosition

				innerLastUsdAmountPosition, err := findAmountEndIndex(text, indices[9]-1)
				if err != nil {
					return true, fmt.Errorf("error finding the end of the USD amount: %w", err)
				}
				lastUsdAmountPosition = innerLastUsdAmountPosition

				return true, nil
			}

			return false, nil
		},
	)
	if err != nil {
		return 0, 0, fmt.Errorf("error finding amounts positions: %w", err)
	}

	return lastArsAmountPosition, lastUsdAmountPosition, nil
}

func extractPastPaymentMovements(rows []*pdf.Row, i int, maxRowIdx int, j int, tablePositions PDFTablePositions) (int, int, []PDFMovement, error) {
	// The first movement should be the previous balance
	prevBalMov, err := extractPreviousBalance(rows, i, maxRowIdx, j)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("error extracting previous balance: %w", err)
	}

	// Find the rest of the movements
	pastPaymentMovements := []PDFMovement{prevBalMov}
	lastI, lastJ, err := processRows(
		rows,
		i,
		maxRowIdx,
		j,
		func(text string) (bool, error) {
			// Find the rest of the past payment movs:
			mov, errMov := extractMovFromText(text, tablePositions)
			if errMov == nil {
				pastPaymentMovements = append(pastPaymentMovements, mov)
			}

			// If we reached the end of the past payment movements, we return the movements we have so far
			if text == movsGroupEndCharacter {
				return true, nil
			}

			return false, nil
		},
	)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("error extracting past payment movements: %w", err)
	}

	// We don't want to return the "end of group" character, so we need to increment the indices
	nextI, nextJ, err := nextIndices(lastI, lastJ, rows)
	if err != nil {
		return 0, 0, nil, fmt.Errorf("error getting next indices: %w", err)
	}

	return nextI, nextJ, pastPaymentMovements, nil
}

func nextIndices(i int, j int, rows []*pdf.Row) (int, int, error) {
	lastJInRow := len(rows[i].Content) - 1
	if j >= lastJInRow {
		i = i + 1
		j = 0
		if i >= len(rows) {
			return 0, 0, fmt.Errorf("can't get next text because i=%d is out of bounds of len(rows)=%d", i, len(rows))
		}

		return i, j, nil
	}

	return i, j + 1, nil

}

func extractMovFromText(text string, tablePositions PDFTablePositions) (PDFMovement, error) {
	rawDate, err := extractRawDate(text, tablePositions)
	if err != nil {
		return PDFMovement{}, fmt.Errorf("error extracting raw date: %w", err)
	}

	arsAmount, firstArsAmountPositionInRow, err := extractArsAmount(text, tablePositions.ARSAmountEnd)
	if err != nil {
		return PDFMovement{}, fmt.Errorf("error extracting ARS amount: %w", err)
	}

	lastDetailPosition := firstArsAmountPositionInRow - 1
	detail, currentInstallment, totalInstallments, err := extractDetailAndInstallments(text, tablePositions, lastDetailPosition)
	if err != nil {
		return PDFMovement{}, fmt.Errorf("error extracting detail and installments: %w", err)
	}

	usdAmount, err := extractUSDAmount(text, tablePositions.USDAmountStart, tablePositions.USDAmountEnd)
	if err != nil {
		return PDFMovement{}, fmt.Errorf("error extracting USD amount: %w", err)
	}

	receipt, err := extractReceipt(text, tablePositions.ReceiptStart, tablePositions.ReceiptEnd)
	if err != nil {
		return PDFMovement{}, fmt.Errorf("error extracting receipt: %w", err)
	}

	dateAsTime, err := VISADotDateToTime(rawDate)
	if err != nil {
		return PDFMovement{}, fmt.Errorf("error converting date %s to time: %w", rawDate, err)
	}

	return PDFMovement{
		OriginalDate:       &dateAsTime,
		ReceiptNumber:      receipt,
		Detail:             detail,
		CurrentInstallment: currentInstallment,
		TotalInstallments:  totalInstallments,
		AmountARS:          arsAmount,
		AmountUSD:          usdAmount,
	}, nil
}

func extractRawDate(text string, tablePositions PDFTablePositions) (string, error) {
	firstPos := tablePositions.OriginalDateStart
	lastPos := tablePositions.OriginalDateEnd + 1
	if firstPos >= len(text) || lastPos >= len(text) || firstPos > lastPos {
		return "", fmt.Errorf(
			"unexpected end of text when extracting raw date. len(text)=%d, firstPos=%d, lastPos=%d",
			len(text),
			firstPos,
			lastPos,
		)
	}
	return strings.TrimSpace(text[firstPos:lastPos]), nil
}

func extractDetailAndInstallments(text string, tablePositions PDFTablePositions, lastDetailPosition int) (string, *int, *int, error) {
	rawDetail := strings.TrimSpace(text[tablePositions.DetailStart : lastDetailPosition+1])
	var currentInstallment *int
	var totalInstallments *int
	matchesDetailWithInstallments := detailWithInstallmentsRegex.FindStringSubmatch(rawDetail)
	if len(matchesDetailWithInstallments) > 1 {
		detail := stringsale.RemoveDuplicateSpaces(strings.TrimSpace(matchesDetailWithInstallments[1]))

		currInst, err := strconv.Atoi(matchesDetailWithInstallments[2])
		if err != nil {
			return "", nil, nil, fmt.Errorf("error converting current installment %s to int: %w", matchesDetailWithInstallments[2], err)
		}
		currentInstallment = pointersale.ToPointer(currInst)

		totInst, err := strconv.Atoi(matchesDetailWithInstallments[3])
		if err != nil {
			return "", nil, nil, fmt.Errorf("error converting total installment %s to int: %w", matchesDetailWithInstallments[3], err)
		}
		totalInstallments = pointersale.ToPointer(totInst)

		return detail, currentInstallment, totalInstallments, nil
	}
	return stringsale.RemoveDuplicateSpaces(rawDetail), currentInstallment, totalInstallments, nil
}

func extractReceipt(text string, start int, end int) (*string, error) {
	if end >= len(text) {
		return nil, fmt.Errorf("unexpected end of text when extracting receipt. len(text)=%d, end=%d", len(text), end)
	}
	rawReceipt := strings.TrimSpace(text[start : end+1])

	if rawReceipt == "" {
		return nil, nil
	}
	return &rawReceipt, nil

}

func extractUSDAmount(text string, start int, end int) (decimal.Decimal, error) {
	if end > len(text) {
		return decimal.Zero, nil
	}
	// If USD amount is negative, then the sign is at the end of the amount. If it's positive, then the position would
	// be out of bounds in the text
	if end == len(text) {
		end--
	}

	sign := ""
	if string(text[end]) == "-" {
		sign = "-"
		end--
	}

	rawUsdAmount := sign + strings.TrimSpace(text[start:end+1])
	asDec, err := PDFAmountToDecimal(rawUsdAmount)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error converting USD amount %s to decimal: %w", rawUsdAmount, err)
	}
	return asDec, nil
}

func extractArsAmount(text string, arsAmountEnd int) (decimal.Decimal, int, error) {
	if arsAmountEnd > len(text) {
		return decimal.Decimal{}, 0, fmt.Errorf("unexpected end of text. len(text)=%d, arsAmountEnd=%d", len(text), arsAmountEnd)
	}

	if arsAmountEnd == 0 {
		return decimal.Decimal{}, 0, fmt.Errorf("unexpected arsAmountEnd=0")
	}
	// If ARS amount is negative, then the sign is at the end of the amount. If it's positive, then the position would
	// be out of bounds in the text
	if arsAmountEnd == len(text) {
		arsAmountEnd--
	}

	lastArsAmountChar := text[arsAmountEnd]

	sign := ""
	arsAmountIndex := arsAmountEnd
	if string(lastArsAmountChar) == "-" {
		sign = "-"
		arsAmountIndex--
	}

	// TODO ALE: this code extracts the table positions for each PDF, but it's always the same. This line that reads the
	//  amount with end - 15 is hardcoded as a hack until I refactor this code to receive the table positions as data when
	//  the extractor is initialized
	arsAmountStart := arsAmountIndex - 15
	rawArsAmount := strings.TrimSpace(text[arsAmountStart : arsAmountIndex+1])

	// Convert final ARS amount to decimal
	finalRawArsAmount := sign + rawArsAmount
	asDec, err := PDFAmountToDecimal(finalRawArsAmount)
	if err != nil {
		return decimal.Decimal{}, 0, fmt.Errorf("error converting ARS amount to decimal: %w", err)
	}
	return asDec, arsAmountStart, nil
}

// extractPreviousBalance is defined separately because it's the only movement that doesn't have a date
func extractPreviousBalance(rows []*pdf.Row, i int, maxRowIdx int, j int) (PDFMovement, error) {
	var mov PDFMovement
	_, _, err := processRows(
		rows,
		i,
		maxRowIdx,
		j,
		func(text string) (bool, error) {
			matchesSaldoAnterior := saldoAnteriorRegex.FindStringSubmatch(text)
			if len(matchesSaldoAnterior) > 1 {
				arsAmount, err := PDFAmountToDecimal(matchesSaldoAnterior[1])
				if err != nil {
					return true, fmt.Errorf("error converting ARS amount %s to decimal: %w", matchesSaldoAnterior[1], err)
				}

				usdAmount, err := PDFAmountToDecimal(matchesSaldoAnterior[2])
				if err != nil {
					return true, fmt.Errorf("error converting USD amount %s to decimal: %w", matchesSaldoAnterior[2], err)
				}

				mov = PDFMovement{
					Detail:    "SALDO ANTERIOR",
					AmountARS: arsAmount,
					AmountUSD: usdAmount,
				}
				return true, nil
			}

			return false, nil
		},
	)
	if err != nil {
		return PDFMovement{}, fmt.Errorf("error extracting previous balance: %w", err)
	}

	return mov, nil
}

func findTableHeadersPosition(rows []*pdf.Row, i int, maxRowIdx int) (int, int, error) {
	lastI, lastJ, err := processRows(
		rows,
		i,
		maxRowIdx,
		0,
		func(text string) (bool, error) {
			if tableHeadersRegex.MatchString(text) {
				return true, nil
			}

			return false, nil
		},
	)
	if err != nil {
		return 0, 0, fmt.Errorf("error finding table headers position: %w", err)
	}

	return lastI, lastJ, nil
}

func extractAllRows(r *pdf.Reader) ([]*pdf.Row, error) {
	allRows := make([]*pdf.Row, 0)
	totalPage := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, err := p.GetTextByRow()
		if err != nil {
			return nil, fmt.Errorf("error getting text by row: %w", err)
		}

		rowsSlice := ([]*pdf.Row)(rows)

		allRows = append(allRows, rowsSlice...)
	}
	return allRows, nil
}

type Document struct {
	TotalARS             decimal.Decimal
	TotalUSD             decimal.Decimal
	CloseDate            time.Time
	ExpirationDate       time.Time
	PastPaymentMovements []PDFMovement
	Cards                []PDFCard
	TablePositions       PDFTablePositions
	TaxesMovements       []PDFMovement
}
