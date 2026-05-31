# Thai Tax & Legal Knowledge Cache

Verified findings cached so agents do not re-research every time. Primarily Thai tax / legal / accounting, but any researched business-domain fact you were unsure about (e.g. how a competitor models a document or flow) belongs here too. Rule that created this file: core-rules "Tax & Accounting Correctness (Research-First)" — trigger is **uncertainty about any domain knowledge**, not only tax features.

> ⚠️ **Tax rules are version-sensitive.** Every entry below is dated. Before shipping a rate/field/form to production or using it for real filing, re-verify the CURRENT value on `rd.go.th`. Royal Decrees and incentive schemes change yearly. Append new dated entries; do not silently overwrite — keep history so we can see when a value changed.

---

## VAT (Value Added Tax)
_As of 2026-05-29:_
- **Statutory rate: 10%** (Revenue Code §80). **Effective rate: 7%** (≈6.3% VAT + 0.7% local tax), applied by Royal Decree and renewed periodically. Latest extension: Royal Decree No. 799 B.E. 2568 — **7% through 30 Sep 2026** (some sources say 1 Oct 2026). Re-verify before that date; it is usually renewed.
- VAT registration threshold and exempt categories: verify on rd.go.th before relying on them (not yet cached).
- Sources: [RD English – Chapter 4 VAT](https://www.rd.go.th/english/37718.html), [Forvis Mazars VAT guide](https://www.forvismazars.com/th/en/insights/doing-business-in-thailand/tax/value-added-tax-vat-in-thailand)

## Full Tax Invoice — required fields (Revenue Code §86/4)
_As of 2026-05-29 — 13 mandatory fields; missing one lets RD deny input-VAT credit:_
1. The word **"ใบกำกับภาษี" / "Tax Invoice"** in a prominent place
2. Seller: **name, address, Taxpayer ID (TIN)**
3. Buyer: **name, address** (+ TIN for full B2B input-credit)
4. **Serial number** of the tax invoice (and book number if any)
5. **Description, type, category, quantity, value** of goods/services
6. **VAT amount**, clearly separated from the goods/services value
7. **Date of issuance**
8. **"สำนักงานใหญ่" / "Head Office" or "สาขาที่ …" / "Branch No. …"** for BOTH seller's and buyer's place of business
- Language/currency: must be **Thai language, THB, Thai/Arabic numerals**. Foreign language/currency needs Director-General approval. Foreign-currency invoices must show both currencies + exchange rate; THB amount is what's reported to RD.
- Maps to our model: branch code `00000` = Head Office (see SKILL §3). Tax-invoice docs must carry seller+buyer branch designation.
- Sources: [RD §85_86](https://www.rd.go.th/english/37741.html), [Thailand tax invoice requirements](https://invoicedataextraction.com/blog/thailand-tax-invoice-requirements)

## Withholding Tax (WHT)
_As of 2026-05-29. Forms: **PND 3** = payee is an individual; **PND 53** = payee is a juristic person (company); PND 1 = salary; PND 54 = foreign company. Rates assume payee has no exemption/treaty relief:_

| Payment type | Rate |
|---|---|
| Professional fees (lawyer, accountant, engineer, consultant) | 3% |
| Services / construction / repair | 3% |
| Rent (office, land, equipment) | 5% |
| Advertising | 2% |
| Transport / logistics | 1% |
| Royalties / IP licensing | 3% |
| Dividends | 10% |
| Bank interest | 1% |
| Prizes / awards | 5% |

- Filing/remittance: within **7 days** of month-end (paper) or **15 days** (e-filing), per Ministry of Finance extension (active 1 Feb 2024 – 31 Jan 2027).
- Penalty: 1.5%/month surcharge + 100–2,000 THB per late form.
- ⚠️ The e-Withholding-Tax rate-reduction incentive **expired 31 Dec 2025** — do not assume reduced e-WHT rates for 2026 unless renewed.
- Sources: [PwC Thailand WHT](https://taxsummaries.pwc.com/thailand/corporate/withholding-taxes), [FlowAccount WHT guide](https://flowaccount.com/blog/withholding-tax-in-thailand/)

## e-Tax Invoice & e-Receipt
_As of 2026-05-29 — currently **voluntary** (not mandatory); paper invoices may run in parallel:_
- **Two systems:**
  - **e-Tax Invoice & e-Receipt** — all business sizes; requires a **digital signature + certificate from an RD-approved CA**.
  - **e-Tax Invoice by Email** — small business with annual revenue **≤ THB 30M**; uses ETDA time-stamp verification.
- Transmit to RD in **XML format within 15 days** of the following month.
- Retain electronic records **≥ 5 years**.
- ⚠️ Easy E-Receipt 2.0 deduction scheme does **not** apply to 2026 spending (no approved scheme for 2026 tax year as of this date).
- Sources: [PKF Thailand e-Tax](https://pkfthailand.asia/understanding-e-tax-invoice-in-thailand/), [VATupdate e-Tax guide](https://www.vatupdate.com/2025/11/11/a-comprehensive-guide-to-e-tax-invoices-in-thailand-types-laws-and-verification-methods/)

---

## How to extend this file
After researching a new tax/accounting topic, append a dated section here with: the verified fact, the date checked, and the source URL(s). If a cached value is now stale, add a new dated line above the old one (keep the old for history) and note the change.
