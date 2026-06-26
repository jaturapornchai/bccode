# ERP Accounting Number Rules

Source: Jead pasted rule, applied to `D:\bccode` on 2026-06-06.

## Scope
Applies to ERP Core, Accounting, Finance, Stock, Sales, Purchase, AR, AP, GL, VAT, Tax, Payment, Cost, Report, and Analytics.

## Iron Rule
Financial correctness beats speed. Do not allow a 0.01 THB mismatch.

Never use floating point for protected accounting fields:
- JavaScript/TypeScript `number`
- Go `float32` / `float64`
- MongoDB double / float64
- PostgreSQL `real`, `double precision`, `float`, `money`
- ClickHouse `Float32`, `Float64`

Protected semantics include amount, price, cost, discount, VAT/tax, debit, credit, balance, total/subtotal, paid/remaining amount, decimal quantity, unit price, average cost, exchange rate, rounding amount, stock value, invoice amount, and payment amount.

## Storage Contract
API may receive and return decimal strings. Databases must store queryable native decimal values.

- MongoDB: `Decimal128`, or `Long` for smallest-unit storage such as satang.
- PostgreSQL: `numeric(P,S)`, or `bigint` for smallest-unit storage.
- ClickHouse: `Decimal(P,S)`, or `Int64` for smallest-unit storage.

Default precision:
- Money: `18,2`
- Unit price: `20,4`
- Quantity: `18,4`
- Exchange rate: `28,8`

## Project Naming
This project also enforces lowercase persisted/API/database names (underscore allowed; snake_case OK — only uppercase/camelCase is a violation). Use project names such as:
- `vatamount`, not `vat_amount`
- `netamount`, not `net_amount`
- `grossamount`, not `gross_amount`
- `unitprice`, not `unit_price`
- `exchangerate`, not `exchange_rate`
- `amountsatang`, not `amount_satang`

External provider fields may keep provider-required names only inside adapter boundaries.

## Code Rules
Frontend and TypeScript calculations must use decimal strings plus a decimal library such as `decimal.js`.

Forbidden in accounting contexts:
- `parseFloat(amount)`
- `Number(amount)`
- unary `+amount`
- `amount * 1.07`
- `price * qty` using JavaScript numbers
- `rows.reduce((sum, row) => sum + Number(row.amount), 0)`

Use decimal library calculations or database decimal aggregation.

## Query Rules
Reports must sum/filter/sort/group directly on decimal fields. Do not cast decimal values to float. Do not create float shadow fields such as `amountfloat`, `totalfloat`, or `balancefloat`.

## Migration Rules
Legacy float/double/number accounting data must not be silently converted. Required migration evidence:
- Backup
- Dry run
- Record count before/after
- Sum totals before/after
- Field conversion log
- Rollback notes
- `NEED_REVIEW_FLOAT_MONEY` markers for suspicious binary-float residue

## Test Rules
Accounting changes must include focused tests where applicable:
- `0.1 + 0.2 = 0.30`
- VAT 7%
- Discount
- Subtotal and net total
- Debit equals credit
- Rounding
- Decimal quantity and unit price
- Exchange rate precision
- API rejects JSON number for amount/price/qty
- MongoDB Decimal128 query, PostgreSQL `SUM(numeric)`, or ClickHouse `sum(Decimal)` for storage/projection changes

## Current Review Findings
Targeted scan on 2026-06-06 found existing legacy violations that must be migrated separately with backup/dry-run/totals evidence, not silently changed:
- `D:\bccode\backend\internal\goapi\myclickhouse\ensure_tables.go` defines many accounting/stock/report fields as `Float64`, including `totalamount`, `price`, `qty`, `discountamount`, `exchangerate`, `averagecost`, `balanceqty`, and `balanceamount`.
- `D:\bccode\frontend\src\lib\product-barcode\types.ts` still models several accounting/stock fields as TypeScript `number`, including `price`, `amount`, `qty`, `averagecost`, and `total`.
- `D:\bccode\frontend\src\app\menu\product-set-screen.tsx` and related product barcode components use `Number(choice.price)` for UI price calculations/display.
- `D:\bccode\backend\internal\goapi\dataimport\xlsx_product.go`, `D:\bccode\backend\internal\goapi\handlers\dataimport\xlsx_product.go`, and `D:\bccode\backend\internal\goapi\handlers\transaction_calculator.go` contain `parseFloat` paths for price/amount/qty/report totals.

Use `NEED_REVIEW_FLOAT_MONEY` when touching these areas until a real migration plan converts storage and calculations to Decimal/native numeric types.
