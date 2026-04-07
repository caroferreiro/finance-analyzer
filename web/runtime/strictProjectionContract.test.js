import test from "node:test";
import assert from "node:assert";

import { buildOverviewProjection } from "./computeTables.js";

const METRICS = [
  "netStatement",
  "cardMovements",
  "newDebt",
  "carryOverDebt",
  "nextMonthDebt",
  "remainingDebt",
  "taxes",
  "pastPayments",
];

function sampleComputeResult() {
  return {
    Tables: [
      {
        TableID: "overview_metrics_by_statement_month",
        Columns: [
          { Key: "statement_month_date", Label: "Statement Month Date", Type: "date" },
          { Key: "net_statement_ars", Label: "Net Statement ARS", Type: "money_ars" },
          { Key: "net_statement_usd", Label: "Net Statement USD", Type: "money_usd" },
          { Key: "card_movements_ars", Label: "Card Movements ARS", Type: "money_ars" },
          { Key: "card_movements_usd", Label: "Card Movements USD", Type: "money_usd" },
          { Key: "new_debt_ars", Label: "New Debt ARS", Type: "money_ars" },
          { Key: "new_debt_usd", Label: "New Debt USD", Type: "money_usd" },
          { Key: "carry_over_debt_ars", Label: "Carry Over Debt ARS", Type: "money_ars" },
          { Key: "carry_over_debt_usd", Label: "Carry Over Debt USD", Type: "money_usd" },
          { Key: "next_month_debt_ars", Label: "Next Month Debt ARS", Type: "money_ars" },
          { Key: "next_month_debt_usd", Label: "Next Month Debt USD", Type: "money_usd" },
          { Key: "remaining_debt_ars", Label: "Remaining Debt ARS", Type: "money_ars" },
          { Key: "remaining_debt_usd", Label: "Remaining Debt USD", Type: "money_usd" },
          { Key: "taxes_ars", Label: "Taxes ARS", Type: "money_ars" },
          { Key: "taxes_usd", Label: "Taxes USD", Type: "money_usd" },
          { Key: "past_payments_ars", Label: "Past Payments ARS", Type: "money_ars" },
          { Key: "past_payments_usd", Label: "Past Payments USD", Type: "money_usd" },
        ],
        Rows: [
          [
            "2025-01-01",
            "8500.00",
            "210.00",
            "7800.00",
            "190.00",
            "5200.00",
            "100.00",
            "2600.00",
            "90.00",
            "1900.00",
            "70.00",
            "3500.00",
            "130.00",
            "800.00",
            "30.00",
            "-100.00",
            "-10.00",
          ],
          [
            "2025-02-01",
            "9200.00",
            "245.00",
            "8400.00",
            "220.00",
            "5800.00",
            "130.00",
            "2600.00",
            "90.00",
            "1700.00",
            "60.00",
            "3300.00",
            "120.00",
            "900.00",
            "35.00",
            "-100.00",
            "-10.00",
          ],
        ],
      },
    ],
  };
}

function legacyFixtureModel() {
  return {
    months: ["2025-01", "2025-02"],
    latestMonth: "2025-02",
    prevMonth: "2025-01",
    latest: {
      currency: {
        ARS: {
          netStatement: 9200,
          cardMovements: 8400,
          newDebt: 5800,
          carryOverDebt: 2600,
          nextMonthDebt: 1700,
          remainingDebt: 3300,
          taxes: 900,
          pastPayments: -100,
        },
        USD: {
          netStatement: 245,
          cardMovements: 220,
          newDebt: 130,
          carryOverDebt: 90,
          nextMonthDebt: 60,
          remainingDebt: 120,
          taxes: 35,
          pastPayments: -10,
        },
      },
    },
    prev: {
      currency: {
        ARS: {
          netStatement: 8500,
          cardMovements: 7800,
          newDebt: 5200,
          carryOverDebt: 2600,
          nextMonthDebt: 1900,
          remainingDebt: 3500,
          taxes: 800,
          pastPayments: -100,
        },
        USD: {
          netStatement: 210,
          cardMovements: 190,
          newDebt: 100,
          carryOverDebt: 90,
          nextMonthDebt: 70,
          remainingDebt: 130,
          taxes: 30,
          pastPayments: -10,
        },
      },
    },
    trend: {
      months: ["2025-01", "2025-02"],
      ARS: {
        netStatement: [8500, 9200],
        cardMovements: [7800, 8400],
        newDebt: [5200, 5800],
        carryOverDebt: [2600, 2600],
        nextMonthDebt: [1900, 1700],
        remainingDebt: [3500, 3300],
        taxes: [800, 900],
        pastPayments: [-100, -100],
      },
      USD: {
        netStatement: [210, 245],
        cardMovements: [190, 220],
        newDebt: [100, 130],
        carryOverDebt: [90, 90],
        nextMonthDebt: [70, 60],
        remainingDebt: [130, 120],
        taxes: [30, 35],
        pastPayments: [-10, -10],
      },
    },
  };
}

function assertCurrencyContract(strictCurrency, legacyCurrency) {
  for (const metric of METRICS) {
    assert.strictEqual(strictCurrency[metric], legacyCurrency[metric], `metric mismatch: ${metric}`);
  }
}

test("strict overview projection matches legacy fixture contract", () => {
  const strict = buildOverviewProjection(sampleComputeResult());
  const legacy = legacyFixtureModel();

  assert.strictEqual(strict.available, true);
  assert.deepStrictEqual(strict.months, legacy.months);
  assert.strictEqual(strict.latestMonth, legacy.latestMonth);
  assert.strictEqual(strict.prevMonth, legacy.prevMonth);

  assertCurrencyContract(strict.latest.currency.ARS, legacy.latest.currency.ARS);
  assertCurrencyContract(strict.latest.currency.USD, legacy.latest.currency.USD);
  assertCurrencyContract(strict.prev.currency.ARS, legacy.prev.currency.ARS);
  assertCurrencyContract(strict.prev.currency.USD, legacy.prev.currency.USD);

  assert.deepStrictEqual(strict.trend.months, legacy.trend.months);
  for (const metric of METRICS) {
    assert.deepStrictEqual(strict.trend.ARS[metric], legacy.trend.ARS[metric], `ARS trend mismatch: ${metric}`);
    assert.deepStrictEqual(strict.trend.USD[metric], legacy.trend.USD[metric], `USD trend mismatch: ${metric}`);
  }
});
