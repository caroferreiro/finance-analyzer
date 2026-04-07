## Demo dataset (public “golden” fixture)


> **Reminder:** Update `docs/index.md` if you add, rename, or remove documents.
### Purpose
- This folder contains a **synthetic/anonymized** dataset that is safe to ship publicly.
- It serves as a single canonical fixture for:
  - Go unit tests (TDD)
  - the hosted static dashboard “demo mode”

### Safety + anonymization checklist
- **Owners**: synthetic labels (`OWNER A`, `OWNER B`), not real names.
- **Merchants / Detail**: synthetic merchant strings (`SUPERMARKET DEMO`, etc.).
- **Card numbers**: fake and obviously not real (`0000`, `1111`, `2222`).
- **No sensitive identifiers**: no addresses, emails, phone numbers, account numbers.

### Files
- `extracted.csv`
  - Semicolon-separated (`;`) to match the CLI output.
  - Header must match the extracted CSV schema exactly.
    - Do not hardcode the full header in this README; it is enforced by Go tests.
- `mappings.v1.json`
  - Minimal exact-match mappings used by the dashboard.

### What this dataset should cover (intentionally, but not rigidly)
- Multiple statement months (at least 2 different `CloseDate` months).
- Installments (at least one `CardMovement` with `CurrentInstallment`/`TotalInstallments` set).
- Refund pairs:
  - one matching positive/negative pair across different months, and
  - one matching positive/negative pair within the same month.
- Mapping gaps (to feed DQ):
  - at least one unmapped category, and
  - at least one unmapped owner.

> Note: Parse-invalid fixtures (bad date formats, invalid MovementType strings, etc.) should live as additional files in this folder (e.g. `extracted.invalid_*.csv`) so the main demo dataset stays usable for the hosted demo.

### Planned simple additions (keep each addition small and purposeful)
- USD examples (at least one `CardMovement` with `AmountUSD != 0`).
- More banks/card companies (additional `CardCompany` / `Bank` combinations). Mercado Pago is now supported by the extractor but not yet represented in the demo dataset.
- Optional-field edge cases (missing receipt/date, blank card fields on non-card movements).
- Whitespace quirks in `Detail` to make normalization value visible.
- A “clean” month with zero DQ issues so the demo looks good.

