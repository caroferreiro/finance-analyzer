package financeengine

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Alechan/finance-analyzer/pkg/internal/pdfcardsummary"
)

// normalizeDetailKey mirrors the JS normalizeDetailKey function:
// trim whitespace, collapse multiple spaces, convert to uppercase.
// This ensures the Go lookup matches the normalized keys built by the JS runtime.
func normalizeDetailKey(detail string) string {
	s := strings.TrimSpace(detail)
	s = strings.Join(strings.Fields(s), " ")
	return strings.ToUpper(s)
}

const (
	DQRuleMissingCategoryMappingForCardMovement = "DQ003"
	DQRuleMissingOwnerMappingForCardMovement    = "DQ004"
)

// CategoryPrefixEntry maps a normalized detail prefix to a category.
// Entries are sorted longest-first so the most specific prefix wins.
type CategoryPrefixEntry struct {
	Prefix   string `json:"prefix"`
	Category string `json:"category"`
}

type Mappings struct {
	OwnersByCardOwner      map[string]string     `json:"ownersByCardOwner"`
	OwnersByCardNumber     map[string]string     `json:"ownersByCardNumber"`
	CategoryByDetail       map[string]string     `json:"categoryByDetail"`
	CategoryByDetailPrefix []CategoryPrefixEntry `json:"categoryByDetailPrefix"`
}

// categoryForDetail returns the mapped category for a detail string,
// trying exact match first and then prefix patterns. Returns "" if unmapped.
func categoryForDetail(detail string, mappings Mappings) string {
	normalized := normalizeDetailKey(detail)
	if mappings.CategoryByDetail != nil {
		if cat, ok := mappings.CategoryByDetail[normalized]; ok {
			return cat
		}
	}
	for _, entry := range mappings.CategoryByDetailPrefix {
		if strings.HasPrefix(normalized, entry.Prefix) {
			return entry.Category
		}
	}
	return ""
}

type DQIssue struct {
	RuleID       string
	Message      string
	MovementType pdfcardsummary.MovementType
	CloseDate    time.Time
	Detail       string
	CardOwner    string
	CardNumber   *string
}

type DQSummaryByRuleRow struct {
	RuleID string
	Count  int
}

func (e *Engine) DataQuality(rows []pdfcardsummary.MovementWithCardContext, mappings Mappings) ([]DQIssue, []DQSummaryByRuleRow) {
	issues := make([]DQIssue, 0)
	countByRule := make(map[string]int)

	for _, row := range rows {
		if row.MovementType != pdfcardsummary.MovementTypeCard || row.CardContext == nil {
			continue
		}

		if !isMappedOwner(row, mappings) {
			issue := DQIssue{
				RuleID:       DQRuleMissingOwnerMappingForCardMovement,
				Message:      fmt.Sprintf("missing owner mapping for CardMovement (owner=%q, card=%q)", row.CardOwner, derefString(row.CardNumber)),
				MovementType: row.MovementType,
				CloseDate:    row.CloseDate,
				Detail:       row.Movement.Detail,
				CardOwner:    row.CardOwner,
				CardNumber:   row.CardNumber,
			}
			issues = append(issues, issue)
			countByRule[issue.RuleID]++
		}

		if !isMappedCategory(row, mappings) {
			issue := DQIssue{
				RuleID:       DQRuleMissingCategoryMappingForCardMovement,
				Message:      fmt.Sprintf("missing category mapping for CardMovement (detail=%q)", row.Movement.Detail),
				MovementType: row.MovementType,
				CloseDate:    row.CloseDate,
				Detail:       row.Movement.Detail,
				CardOwner:    row.CardOwner,
				CardNumber:   row.CardNumber,
			}
			issues = append(issues, issue)
			countByRule[issue.RuleID]++
		}
	}

	slices.SortFunc(issues, func(a, b DQIssue) int {
		if a.RuleID < b.RuleID {
			return -1
		}
		if a.RuleID > b.RuleID {
			return 1
		}
		if a.CloseDate.Before(b.CloseDate) {
			return -1
		}
		if a.CloseDate.After(b.CloseDate) {
			return 1
		}
		if a.Detail < b.Detail {
			return -1
		}
		if a.Detail > b.Detail {
			return 1
		}
		if a.CardOwner < b.CardOwner {
			return -1
		}
		if a.CardOwner > b.CardOwner {
			return 1
		}
		return 0
	})

	summary := make([]DQSummaryByRuleRow, 0, len(countByRule))
	for ruleID, count := range countByRule {
		summary = append(summary, DQSummaryByRuleRow{
			RuleID: ruleID,
			Count:  count,
		})
	}
	slices.SortFunc(summary, func(a, b DQSummaryByRuleRow) int {
		if a.RuleID < b.RuleID {
			return -1
		}
		if a.RuleID > b.RuleID {
			return 1
		}
		return 0
	})

	return issues, summary
}

func isMappedOwner(row pdfcardsummary.MovementWithCardContext, mappings Mappings) bool {
	if mappings.OwnersByCardOwner != nil {
		if _, ok := mappings.OwnersByCardOwner[strings.ToUpper(row.CardOwner)]; ok {
			return true
		}
	}

	if mappings.OwnersByCardNumber != nil && row.CardNumber != nil {
		if _, ok := mappings.OwnersByCardNumber[*row.CardNumber]; ok {
			return true
		}
	}

	return false
}

func isMappedCategory(row pdfcardsummary.MovementWithCardContext, mappings Mappings) bool {
	return categoryForDetail(row.Movement.Detail, mappings) != ""
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
