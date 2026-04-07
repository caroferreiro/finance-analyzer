package santander

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_expirationDateRegex(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		shouldMatch   bool
		expectedGroup string // The captured group if it matches
	}{
		{
			name:          "empty string",
			text:          "",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "non-empty with no matches",
			text:          "a string that doesn't match anything",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "valid match with single space",
			text:          "VENCIMIENTO 01 Dic 24",
			shouldMatch:   true,
			expectedGroup: "01 Dic 24",
		},
		{
			name:          "valid match with multiple spaces",
			text:          "VENCIMIENTO  01 Dic 24",
			shouldMatch:   true,
			expectedGroup: "01 Dic 24",
		},
		{
			name:          "valid match with day starting with 0",
			text:          "VENCIMIENTO 02 Dic 24",
			shouldMatch:   true,
			expectedGroup: "02 Dic 24",
		},
		{
			name:          "valid match with year starting with 0",
			text:          "VENCIMIENTO 01 Dic 04",
			shouldMatch:   true,
			expectedGroup: "01 Dic 04",
		},
		{
			name:          "invalid - no VENCIMIENTO prefix",
			text:          "01 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - wrong prefix",
			text:          "VENCIMIENTOX 01 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - day has more than 2 digits",
			text:          "VENCIMIENTO 123 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - month has more than 3 letters",
			text:          "VENCIMIENTO 01 Diciem 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - year has more than 2 digits",
			text:          "VENCIMIENTO 01 Dic 244",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - has text after the date",
			text:          "VENCIMIENTO 01 Dic 24 extra text",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - has text before VENCIMIENTO",
			text:          "some text VENCIMIENTO 01 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			defaultCfg := DefaultConfig()

			// When
			matches := defaultCfg.ExpirationDateRegex.FindStringSubmatch(tt.text)

			// Then
			if tt.shouldMatch {
				require.NotNil(t, matches, "Expected text to match but it didn't")
				require.Len(t, matches, 2, "Expected 2 groups (full match and captured group)")
				require.Equal(t, tt.text, matches[0], "Full match should be the entire text")
				require.Equal(t, tt.expectedGroup, matches[1], "Captured group should match expected")
			} else {
				require.Nil(t, matches, "Expected text to not match but it did")
			}
		})
	}
}

func Test_closingDateRegex(t *testing.T) {
	tests := []struct {
		name          string
		text          string
		shouldMatch   bool
		expectedGroup string // The captured group if it matches
	}{
		{
			name:          "empty string",
			text:          "",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "non-empty with no matches",
			text:          "a string that doesn't match anything",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "valid match with single space",
			text:          "CIERRE 01 Dic 24",
			shouldMatch:   true,
			expectedGroup: "01 Dic 24",
		},
		{
			name:          "valid match with multiple spaces",
			text:          "CIERRE  01 Dic 24",
			shouldMatch:   true,
			expectedGroup: "01 Dic 24",
		},
		{
			name:          "valid match with day starting with 0",
			text:          "CIERRE 02 Dic 24",
			shouldMatch:   true,
			expectedGroup: "02 Dic 24",
		},
		{
			name:          "valid match with year starting with 0",
			text:          "CIERRE 01 Dic 04",
			shouldMatch:   true,
			expectedGroup: "01 Dic 04",
		},
		{
			name:          "invalid - no CIERRE prefix",
			text:          "01 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - wrong prefix",
			text:          "CIERREX 01 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - day has more than 2 digits",
			text:          "CIERRE 123 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - month has more than 3 letters",
			text:          "CIERRE 01 Diciem 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - year has more than 2 digits",
			text:          "CIERRE 01 Dic 244",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - has text after the date",
			text:          "CIERRE 01 Dic 24 extra text",
			shouldMatch:   false,
			expectedGroup: "",
		},
		{
			name:          "invalid - has text before CIERRE",
			text:          "some text CIERRE 01 Dic 24",
			shouldMatch:   false,
			expectedGroup: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			defaultCfg := DefaultConfig()

			// When
			matches := defaultCfg.ClosingDateRegex.FindStringSubmatch(tt.text)

			// Then
			if tt.shouldMatch {
				require.NotNil(t, matches, "Expected text to match but it didn't")
				require.Len(t, matches, 2, "Expected 2 groups (full match and captured group)")
				require.Equal(t, tt.text, matches[0], "Full match should be the entire text")
				require.Equal(t, tt.expectedGroup, matches[1], "Captured group should match expected")
			} else {
				require.Nil(t, matches, "Expected text to not match but it did")
			}
		})
	}
}

func Test_saldoAnteriorRegex(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		shouldMatch    bool
		expectedGroups []string // The captured groups if it matches
	}{
		{
			name:           "empty string",
			text:           "",
			shouldMatch:    false,
			expectedGroups: nil,
		},
		{
			name:           "non-empty with no matches",
			text:           "a string that doesn't match anything",
			shouldMatch:    false,
			expectedGroups: nil,
		},
		{
			name:           "valid match with single space",
			text:           "SALDO ANTERIOR 660.914,93 832,22",
			shouldMatch:    true,
			expectedGroups: []string{"660.914,93", "832,22"},
		},
		{
			name:           "valid match with multiple spaces",
			text:           "SALDO ANTERIOR  1278.117,92  1,20-",
			shouldMatch:    true,
			expectedGroups: []string{"1278.117,92", "1,20-"},
		},
		{
			name:           "valid match with leading spaces",
			text:           "   SALDO ANTERIOR 255.497,38 22,77",
			shouldMatch:    true,
			expectedGroups: []string{"255.497,38", "22,77"},
		},
		{
			name:           "valid match with zero USD amount",
			text:           "SALDO ANTERIOR 775.125,95 0,00",
			shouldMatch:    true,
			expectedGroups: []string{"775.125,95", "0,00"},
		},
		{
			name:           "valid match with large numbers",
			text:           "SALDO ANTERIOR 1.589.436,69 0,83-",
			shouldMatch:    true,
			expectedGroups: []string{"1.589.436,69", "0,83-"},
		},
		{
			name:           "valid match with multiple spaces between SALDO and ANTERIOR",
			text:           "                        SALDO ANTERIOR                                           660.914,93             832,22          ",
			shouldMatch:    true,
			expectedGroups: []string{"660.914,93", "832,22"},
		},
		{
			name:           "invalid - no SALDO ANTERIOR prefix",
			text:           "660.914,93 832,22",
			shouldMatch:    false,
			expectedGroups: nil,
		},
		{
			name:           "invalid - wrong prefix",
			text:           "SALDO ANTERIORX 660.914,93 832,22",
			shouldMatch:    false,
			expectedGroups: nil,
		},
		{
			name:           "invalid - missing USD amount",
			text:           "SALDO ANTERIOR 660.914,93",
			shouldMatch:    false,
			expectedGroups: nil,
		},
		{
			name:           "invalid - missing ARS amount",
			text:           "SALDO ANTERIOR 832,22",
			shouldMatch:    false,
			expectedGroups: nil,
		},
		{
			name:           "invalid - has text after amounts",
			text:           "SALDO ANTERIOR 660.914,93 832,22 extra text",
			shouldMatch:    false,
			expectedGroups: nil,
		},
		{
			name:           "invalid - has text before SALDO ANTERIOR",
			text:           "some text SALDO ANTERIOR 660.914,93 832,22",
			shouldMatch:    false,
			expectedGroups: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			defaultCfg := DefaultConfig()

			// When
			matches := defaultCfg.SaldoAnteriorRegex.FindStringSubmatch(tt.text)

			// Then
			if tt.shouldMatch {
				require.NotNil(t, matches, "Expected text to match but it didn't")
				require.Len(t, matches, 3, "Expected 3 groups (full match and 2 captured groups)")
				require.Equal(t, tt.text, matches[0], "Full match should be the entire text")
				require.Equal(t, tt.expectedGroups[0], matches[1], "First captured group (ARS amount) should match expected")
				require.Equal(t, tt.expectedGroups[1], matches[2], "Second captured group (USD amount) should match expected")
			} else {
				require.Nil(t, matches, "Expected text to not match but it did")
			}
		})
	}
}

func Test_amountOnlyRowRegex(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		shouldMatch bool
	}{
		{
			name:        "empty string",
			text:        "",
			shouldMatch: false,
		},
		{
			name:        "non-empty with no matches",
			text:        "a string that doesn't match anything",
			shouldMatch: false,
		},
		{
			name:        "broken line - amount only row with ARS amount",
			text:        "                                                                    1.316",
			shouldMatch: true,
		},
		{
			name:        "broken line - amount only row with ARS amount and leading spaces",
			text:        "                                                                    1.316,68",
			shouldMatch: true,
		},
		{
			name:        "broken line - amount only row with large ARS amount",
			text:        "                                                                    12.345,67",
			shouldMatch: true,
		},
		{
			name:        "broken line - amount only row with ARS amount and trailing spaces",
			text:        "                                                                    1.316                    ",
			shouldMatch: true,
		},
		{
			name:        "correct line without date - should NOT match (has receipt and detail)",
			text:        "           28 004170 *  MERCADOPAGO*PAGOCREDI3271                                  5.704,45",
			shouldMatch: false,
		},
		{
			name:        "correct line with full date - should NOT match (has date, receipt, detail)",
			text:        "24 Diciem. 25 222222 *  MERCHANT*PAYMENT         99999999999                   1.234,56",
			shouldMatch: false,
		},
		{
			name:        "row with only whitespace - should NOT match (no amount)",
			text:        "                                                                                    ",
			shouldMatch: false,
		},
		{
			name:        "row with date but no amount - should NOT match (has date)",
			text:        "           01                                                                        ",
			shouldMatch: false,
		},
		{
			name:        "row with receipt but no amount - should NOT match (has receipt)",
			text:        "                    111111 *                                                         ",
			shouldMatch: false,
		},
		{
			name:        "row with detail but no amount - should NOT match (has detail)",
			text:        "                                    TEST MERCHANT                                    ",
			shouldMatch: false,
		},
		{
			name:        "broken line - amount only row with negative ARS amount",
			text:        "                                                                    1.316,68-",
			shouldMatch: true,
		},
		{
			name:        "broken line - amount only row with USD amount",
			text:        "                                                                                    54,68",
			shouldMatch: true,
		},
		{
			name:        "broken line - amount only row with negative USD amount",
			text:        "                                                                                    54,68-",
			shouldMatch: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given
			defaultCfg := DefaultConfig()

			// When
			matches := defaultCfg.AmountOnlyRowRegex.FindStringSubmatch(tt.text)

			// Then
			if tt.shouldMatch {
				require.NotNil(t, matches, "Expected text to match but it didn't")
			} else {
				require.Nil(t, matches, "Expected text to not match but it did")
			}
		})
	}
}
