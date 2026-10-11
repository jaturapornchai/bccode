package report

import (
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

const amountScale int32 = 2

// RoundAmount rounds one stored decimal value once, half away from zero, at scale (empty/invalid = "0.00" style zero)
func RoundAmount(value string, scale int32) string {
	d, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return decimal.Zero.StringFixed(scale)
	}
	return d.StringFixed(scale)
}

// grandTotals keeps the whitelisted totals of the source: amounts rounded once at scale 2, counts as integers
func grandTotals(spec []TotalKey, raw map[string]string) map[string]string {
	out := map[string]string{}
	for _, t := range spec {
		value, ok := raw[t.Key]
		if !ok {
			continue
		}
		if t.Int {
			out[t.Key] = countText(value)
		} else {
			out[t.Key] = RoundAmount(value, amountScale)
		}
	}
	return out
}

func countText(value string) string {
	n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(n, 10)
}

// PageTotal = the sums of one printed page (Champ "ยอดรวม/ต่อหน้า") and the running total through that page
// ("ยอดยกไป", printed again as "ยอดยกมา" at the top of the next page). Rows are 1-based and inclusive.
type PageTotal struct {
	PrintPage int               `json:"print_page"`
	FirstRow  int               `json:"first_row"`
	LastRow   int               `json:"last_row"`
	Sums      map[string]string `json:"sums"`
	Carried   map[string]string `json:"carried"`
}

// pageTotals splits rows into printed pages of rowsPerPage lines and sums the Page totals over detail rows only
// (subtotal/opening/control rows would count twice). Count totals count the detail rows. Sums are exact decimals,
// rounded once when rendered.
func pageTotals(rows []map[string]string, spec []TotalKey, rowsPerPage int) []PageTotal {
	keys := []TotalKey{}
	for _, t := range spec {
		if t.Page {
			keys = append(keys, t)
		}
	}
	out := []PageTotal{}
	if len(keys) == 0 || rowsPerPage <= 0 {
		return out
	}
	running := map[string]decimal.Decimal{}
	for start := 0; start < len(rows); start += rowsPerPage {
		end := start + rowsPerPage
		if end > len(rows) {
			end = len(rows)
		}
		sums := map[string]decimal.Decimal{}
		for _, row := range rows[start:end] {
			if row[KindKey] != KindDetail {
				continue
			}
			for _, t := range keys {
				if t.Int {
					sums[t.Key] = sums[t.Key].Add(decimal.NewFromInt(1))
					continue
				}
				if d, err := decimal.NewFromString(strings.TrimSpace(row[t.Key])); err == nil {
					sums[t.Key] = sums[t.Key].Add(d)
				}
			}
		}
		page := PageTotal{PrintPage: len(out) + 1, FirstRow: start + 1, LastRow: end, Sums: map[string]string{}, Carried: map[string]string{}}
		for _, t := range keys {
			running[t.Key] = running[t.Key].Add(sums[t.Key])
			page.Sums[t.Key] = renderTotal(sums[t.Key], t.Int)
			page.Carried[t.Key] = renderTotal(running[t.Key], t.Int)
		}
		out = append(out, page)
	}
	return out
}

func renderTotal(d decimal.Decimal, isInt bool) string {
	if isInt {
		return d.StringFixed(0)
	}
	return d.StringFixed(amountScale)
}
