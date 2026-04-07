package financeengine

import (
	"slices"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
	"github.com/shopspring/decimal"
)

type OverviewByStatementMonthRow struct {
	StatementMonthDate   time.Time
	CardMovementTotalARS decimal.Decimal
	CardMovementTotalUSD decimal.Decimal
}

type OverviewMetricsByStatementMonthRow struct {
	StatementMonthDate time.Time

	NetStatementARS  decimal.Decimal
	NetStatementUSD  decimal.Decimal
	CardMovementsARS decimal.Decimal
	CardMovementsUSD decimal.Decimal

	NewDebtARS       decimal.Decimal
	NewDebtUSD       decimal.Decimal
	CarryOverDebtARS decimal.Decimal
	CarryOverDebtUSD decimal.Decimal
	NextMonthDebtARS decimal.Decimal
	NextMonthDebtUSD decimal.Decimal
	RemainingDebtARS decimal.Decimal
	RemainingDebtUSD decimal.Decimal
	TaxesARS         decimal.Decimal
	TaxesUSD         decimal.Decimal
	PastPaymentsARS  decimal.Decimal
	PastPaymentsUSD  decimal.Decimal
}

type SpendByOwnerRow struct {
	Owner                string
	Month                time.Time
	CardMovementTotalARS decimal.Decimal
	CardMovementTotalUSD decimal.Decimal
}

type SpendByCategoryRow struct {
	Month                time.Time
	Category             string
	CardMovementTotalARS decimal.Decimal
	CardMovementTotalUSD decimal.Decimal
}

type RawExplorerRow struct {
	CardStatementCloseDate time.Time
	CardStatementDueDate   time.Time
	Bank                   string
	CardCompany            string
	MovementDate           *time.Time
	CardNumber             string
	CardOwner              string
	MovementType           string
	ReceiptNumber          string
	Detail                 string
	InstallmentCurrent     *int
	InstallmentTotal       *int
	AmountARS              decimal.Decimal
	AmountUSD              decimal.Decimal
}

type DebtMaturityScheduleByMonthRow struct {
	BaseStatementMonthDate time.Time
	MaturityMonthDate      time.Time
	MonthOffset            int
	InstallmentCount       int
	MaturityTotalARS       decimal.Decimal
	MaturityTotalUSD       decimal.Decimal
}

type MetaSummary struct {
	RowCount            int
	StatementMonthMin   time.Time
	StatementMonthMax   time.Time
	StatementMonthCount int
}

type Engine struct{}

func New() *Engine {
	return &Engine{}
}

func (e *Engine) OverviewByStatementMonth(rows []pdfcardsummary.MovementWithCardContext) []OverviewByStatementMonthRow {
	byMonth := make(map[time.Time]OverviewByStatementMonthRow)

	for _, row := range rows {
		if row.MovementType != pdfcardsummary.MovementTypeCard {
			continue
		}

		month := time.Date(row.CloseDate.Year(), row.CloseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		acc := byMonth[month]
		acc.StatementMonthDate = month
		acc.CardMovementTotalARS = acc.CardMovementTotalARS.Add(row.Movement.AmountARS)
		acc.CardMovementTotalUSD = acc.CardMovementTotalUSD.Add(row.Movement.AmountUSD)
		byMonth[month] = acc
	}

	result := make([]OverviewByStatementMonthRow, 0, len(byMonth))
	for _, row := range byMonth {
		result = append(result, row)
	}

	slices.SortFunc(result, func(a, b OverviewByStatementMonthRow) int {
		return a.StatementMonthDate.Compare(b.StatementMonthDate)
	})

	return result
}

func (e *Engine) OverviewMetricsByStatementMonth(rows []pdfcardsummary.MovementWithCardContext) []OverviewMetricsByStatementMonthRow {
	byMonth := make(map[time.Time]OverviewMetricsByStatementMonthRow)

	for _, row := range rows {
		month := time.Date(row.CloseDate.Year(), row.CloseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		acc := byMonth[month]
		acc.StatementMonthDate = month

		amountARS := row.Movement.AmountARS
		amountUSD := row.Movement.AmountUSD

		acc.NetStatementARS = acc.NetStatementARS.Add(amountARS)
		acc.NetStatementUSD = acc.NetStatementUSD.Add(amountUSD)

		switch row.MovementType {
		case pdfcardsummary.MovementTypeCard:
			acc.CardMovementsARS = acc.CardMovementsARS.Add(amountARS)
			acc.CardMovementsUSD = acc.CardMovementsUSD.Add(amountUSD)

			currentInstallment, totalInstallments, hasInstallment := validInstallment(row.Movement)
			if hasInstallment {
				if currentInstallment <= 1 {
					acc.NewDebtARS = acc.NewDebtARS.Add(amountARS)
					acc.NewDebtUSD = acc.NewDebtUSD.Add(amountUSD)
				} else {
					acc.CarryOverDebtARS = acc.CarryOverDebtARS.Add(amountARS)
					acc.CarryOverDebtUSD = acc.CarryOverDebtUSD.Add(amountUSD)
				}

				if currentInstallment < totalInstallments {
					acc.NextMonthDebtARS = acc.NextMonthDebtARS.Add(amountARS)
					acc.NextMonthDebtUSD = acc.NextMonthDebtUSD.Add(amountUSD)

					remainingInstallments := decimal.NewFromInt(int64(totalInstallments - currentInstallment))
					acc.RemainingDebtARS = acc.RemainingDebtARS.Add(amountARS.Mul(remainingInstallments))
					acc.RemainingDebtUSD = acc.RemainingDebtUSD.Add(amountUSD.Mul(remainingInstallments))
				}
			} else {
				acc.NewDebtARS = acc.NewDebtARS.Add(amountARS)
				acc.NewDebtUSD = acc.NewDebtUSD.Add(amountUSD)
			}
		case pdfcardsummary.MovementTypeTax:
			acc.TaxesARS = acc.TaxesARS.Add(amountARS)
			acc.TaxesUSD = acc.TaxesUSD.Add(amountUSD)
		case pdfcardsummary.MovementTypePastPayment:
			acc.PastPaymentsARS = acc.PastPaymentsARS.Add(amountARS)
			acc.PastPaymentsUSD = acc.PastPaymentsUSD.Add(amountUSD)
		}

		byMonth[month] = acc
	}

	result := make([]OverviewMetricsByStatementMonthRow, 0, len(byMonth))
	for _, row := range byMonth {
		result = append(result, row)
	}

	slices.SortFunc(result, func(a, b OverviewMetricsByStatementMonthRow) int {
		return a.StatementMonthDate.Compare(b.StatementMonthDate)
	})

	return result
}

func validInstallment(movement pdfcardsummary.Movement) (int, int, bool) {
	if movement.CurrentInstallment == nil || movement.TotalInstallments == nil {
		return 0, 0, false
	}

	currentInstallment := *movement.CurrentInstallment
	totalInstallments := *movement.TotalInstallments
	if currentInstallment <= 0 || totalInstallments <= 0 || totalInstallments < currentInstallment {
		return 0, 0, false
	}

	return currentInstallment, totalInstallments, true
}

func (e *Engine) DebtMaturityScheduleByMonth(rows []pdfcardsummary.MovementWithCardContext) []DebtMaturityScheduleByMonthRow {
	if len(rows) == 0 {
		return nil
	}

	latestStatementMonth := time.Time{}
	for i, row := range rows {
		statementMonth := time.Date(row.CloseDate.Year(), row.CloseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		if i == 0 || statementMonth.After(latestStatementMonth) {
			latestStatementMonth = statementMonth
		}
	}

	type maturityBucket struct {
		monthOffset      int
		installmentCount int
		totalARS         decimal.Decimal
		totalUSD         decimal.Decimal
	}

	byMaturityMonth := make(map[time.Time]maturityBucket)
	for _, row := range rows {
		if row.MovementType != pdfcardsummary.MovementTypeCard {
			continue
		}

		statementMonth := time.Date(row.CloseDate.Year(), row.CloseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		if !statementMonth.Equal(latestStatementMonth) {
			continue
		}

		currentInstallment, totalInstallments, hasInstallment := validInstallment(row.Movement)
		if !hasInstallment {
			continue
		}

		remainingInstallments := totalInstallments - currentInstallment
		if remainingInstallments <= 0 {
			continue
		}

		for monthOffset := 1; monthOffset <= remainingInstallments; monthOffset++ {
			maturityMonth := latestStatementMonth.AddDate(0, monthOffset, 0)
			acc := byMaturityMonth[maturityMonth]
			acc.monthOffset = monthOffset
			acc.installmentCount++
			acc.totalARS = acc.totalARS.Add(row.Movement.AmountARS)
			acc.totalUSD = acc.totalUSD.Add(row.Movement.AmountUSD)
			byMaturityMonth[maturityMonth] = acc
		}
	}

	result := make([]DebtMaturityScheduleByMonthRow, 0, len(byMaturityMonth))
	for maturityMonth, bucket := range byMaturityMonth {
		result = append(result, DebtMaturityScheduleByMonthRow{
			BaseStatementMonthDate: latestStatementMonth,
			MaturityMonthDate:      maturityMonth,
			MonthOffset:            bucket.monthOffset,
			InstallmentCount:       bucket.installmentCount,
			MaturityTotalARS:       bucket.totalARS,
			MaturityTotalUSD:       bucket.totalUSD,
		})
	}

	slices.SortFunc(result, func(a, b DebtMaturityScheduleByMonthRow) int {
		if c := a.MaturityMonthDate.Compare(b.MaturityMonthDate); c != 0 {
			return c
		}
		if a.MonthOffset < b.MonthOffset {
			return -1
		}
		if a.MonthOffset > b.MonthOffset {
			return 1
		}
		return 0
	})

	return result
}

func (e *Engine) SpendByOwner(rows []pdfcardsummary.MovementWithCardContext) []SpendByOwnerRow {
	type key struct {
		owner string
		month time.Time
	}
	byOwnerAndMonth := make(map[key]SpendByOwnerRow)

	for _, row := range rows {
		if row.MovementType != pdfcardsummary.MovementTypeCard || row.CardContext == nil {
			continue
		}

		owner := row.CardOwner
		month := time.Date(row.CloseDate.Year(), row.CloseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		k := key{owner: owner, month: month}
		acc := byOwnerAndMonth[k]
		acc.Owner = owner
		acc.Month = month
		acc.CardMovementTotalARS = acc.CardMovementTotalARS.Add(row.Movement.AmountARS)
		acc.CardMovementTotalUSD = acc.CardMovementTotalUSD.Add(row.Movement.AmountUSD)
		byOwnerAndMonth[k] = acc
	}

	result := make([]SpendByOwnerRow, 0, len(byOwnerAndMonth))
	for _, row := range byOwnerAndMonth {
		result = append(result, row)
	}

	slices.SortFunc(result, func(a, b SpendByOwnerRow) int {
		if c := b.Month.Compare(a.Month); c != 0 {
			return c
		}
		if a.Owner < b.Owner {
			return -1
		}
		if a.Owner > b.Owner {
			return 1
		}
		return 0
	})

	return result
}

func (e *Engine) SpendByCategory(rows []pdfcardsummary.MovementWithCardContext, mappings Mappings) []SpendByCategoryRow {
	type key struct {
		month    time.Time
		category string
	}
	byMonthAndCategory := make(map[key]SpendByCategoryRow)

	for _, row := range rows {
		if row.MovementType != pdfcardsummary.MovementTypeCard {
			continue
		}

		month := time.Date(row.CloseDate.Year(), row.CloseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		category := categoryForDetail(row.Movement.Detail, mappings)
		if category == "" {
			category = "Uncategorized"
		}
		k := key{month: month, category: category}
		acc := byMonthAndCategory[k]
		acc.Month = month
		acc.Category = category
		acc.CardMovementTotalARS = acc.CardMovementTotalARS.Add(row.Movement.AmountARS)
		acc.CardMovementTotalUSD = acc.CardMovementTotalUSD.Add(row.Movement.AmountUSD)
		byMonthAndCategory[k] = acc
	}

	result := make([]SpendByCategoryRow, 0, len(byMonthAndCategory))
	for _, row := range byMonthAndCategory {
		result = append(result, row)
	}

	slices.SortFunc(result, func(a, b SpendByCategoryRow) int {
		if c := b.Month.Compare(a.Month); c != 0 {
			return c
		}
		if a.Category < b.Category {
			return -1
		}
		if a.Category > b.Category {
			return 1
		}
		return 0
	})

	return result
}

func (e *Engine) RawExplorerRows(rows []pdfcardsummary.MovementWithCardContext) []RawExplorerRow {
	result := make([]RawExplorerRow, 0, len(rows))

	for _, row := range rows {
		cardNumber := ""
		cardOwner := ""
		if row.CardContext != nil {
			cardOwner = row.CardContext.CardOwner
			if row.CardContext.CardNumber != nil {
				cardNumber = *row.CardContext.CardNumber
			}
		}

		receiptNumber := ""
		if row.Movement.ReceiptNumber != nil {
			receiptNumber = *row.Movement.ReceiptNumber
		}

		result = append(result, RawExplorerRow{
			CardStatementCloseDate: row.CloseDate,
			CardStatementDueDate:   row.ExpirationDate,
			Bank:                   string(row.Bank),
			CardCompany:            string(row.CardCompany),
			MovementDate:           row.Movement.OriginalDate,
			CardNumber:             cardNumber,
			CardOwner:              cardOwner,
			MovementType:           string(row.MovementType),
			ReceiptNumber:          receiptNumber,
			Detail:                 row.Movement.Detail,
			InstallmentCurrent:     row.Movement.CurrentInstallment,
			InstallmentTotal:       row.Movement.TotalInstallments,
			AmountARS:              row.Movement.AmountARS,
			AmountUSD:              row.Movement.AmountUSD,
		})
	}

	slices.SortFunc(result, func(a, b RawExplorerRow) int {
		if c := b.CardStatementCloseDate.Compare(a.CardStatementCloseDate); c != 0 {
			return c
		}
		if c := b.CardStatementDueDate.Compare(a.CardStatementDueDate); c != 0 {
			return c
		}
		if c := compareTimePtrDesc(a.MovementDate, b.MovementDate); c != 0 {
			return c
		}
		if a.Bank < b.Bank {
			return -1
		}
		if a.Bank > b.Bank {
			return 1
		}
		if a.CardCompany < b.CardCompany {
			return -1
		}
		if a.CardCompany > b.CardCompany {
			return 1
		}
		if a.CardOwner < b.CardOwner {
			return -1
		}
		if a.CardOwner > b.CardOwner {
			return 1
		}
		if a.CardNumber < b.CardNumber {
			return -1
		}
		if a.CardNumber > b.CardNumber {
			return 1
		}
		if a.MovementType < b.MovementType {
			return -1
		}
		if a.MovementType > b.MovementType {
			return 1
		}
		if a.ReceiptNumber < b.ReceiptNumber {
			return -1
		}
		if a.ReceiptNumber > b.ReceiptNumber {
			return 1
		}
		if a.Detail < b.Detail {
			return -1
		}
		if a.Detail > b.Detail {
			return 1
		}
		if c := a.AmountARS.Cmp(b.AmountARS); c != 0 {
			return c
		}
		return a.AmountUSD.Cmp(b.AmountUSD)
	})

	return result
}

func compareTimePtrDesc(a, b *time.Time) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return 1
	}
	if b == nil {
		return -1
	}
	return b.Compare(*a)
}

func (e *Engine) MetaSummary(rows []pdfcardsummary.MovementWithCardContext) MetaSummary {
	if len(rows) == 0 {
		return MetaSummary{}
	}

	monthSet := make(map[time.Time]struct{})
	var minMonth time.Time
	var maxMonth time.Time
	for i, row := range rows {
		month := time.Date(row.CloseDate.Year(), row.CloseDate.Month(), 1, 0, 0, 0, 0, time.UTC)
		monthSet[month] = struct{}{}

		if i == 0 || month.Before(minMonth) {
			minMonth = month
		}
		if i == 0 || month.After(maxMonth) {
			maxMonth = month
		}
	}

	return MetaSummary{
		RowCount:            len(rows),
		StatementMonthMin:   minMonth,
		StatementMonthMax:   maxMonth,
		StatementMonthCount: len(monthSet),
	}
}
