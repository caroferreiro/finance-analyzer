package testsale

import (
	"testing"

	"github.com/shopspring/decimal"
	"go.uber.org/mock/gomock"
)

func TestAssertDecimalEqual(t *testing.T) {
	t.Run("equal decimals pass", func(t *testing.T) {
		// Given
		expected := decimal.NewFromFloat(123.45)
		actual := decimal.NewFromFloat(123.45)

		// When & Then
		AssertDecimalEqual(t, expected, actual)
	})

	t.Run("unequal decimals fail with clear message", func(t *testing.T) {
		// Given
		ctrl := gomock.NewController(t)
		mockT := NewMockTestingTB(ctrl)
		expected := decimal.NewFromFloat(123.45)
		actual := decimal.NewFromFloat(123.46)
		mockT.EXPECT().Helper()
		mockT.EXPECT().Fatalf("decimal values do not match:\nexpected: %s\nactual:   %s", expected.String(), actual.String())

		// When & Then
		AssertDecimalEqual(mockT, expected, actual)
	})

	t.Run("handles zero values", func(t *testing.T) {
		// Given
		expected := decimal.Zero
		actual := decimal.Zero

		// When & Then
		AssertDecimalEqual(t, expected, actual)
	})

	t.Run("handles negative values", func(t *testing.T) {
		// Given
		expected := decimal.NewFromFloat(-123.45)
		actual := decimal.NewFromFloat(-123.45)

		// When & Then
		AssertDecimalEqual(t, expected, actual)
	})

	t.Run("handles large values", func(t *testing.T) {
		// Given
		expected := decimal.NewFromFloat(123456789.123456789)
		actual := decimal.NewFromFloat(123456789.123456789)

		// When & Then
		AssertDecimalEqual(t, expected, actual)
	})
}
