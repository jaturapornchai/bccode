package report

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/shopspring/decimal"

	"smlcloudplatform/internal/goapi/language"
)

// formulaStart = cell text a spreadsheet would run as a formula; such text cells get a leading apostrophe
var formulaStart = regexp.MustCompile(`^\s*[=+\-@\t\r]`)

// glRowLabels = captions of GL total/opening rows in the description column, the same text the screen shows
// (frontend gl-reports.tsx controlRowText)
var glRowLabels = map[string]string{"opening": "gl_opening_balance_2", "carry": "gl_closing_balance", "total": "gl_total", "daytotal": "gl_report_day_total"}

// cellText = what one cell of a row prints: row captions of the GL engine and master lists, raw values otherwise
func cellText(row map[string]string, key string, first bool, lang string) string {
	value := row[key]
	if label := row[LabelKey]; label != "" && first {
		return label
	}
	switch rowtype := row["rowtype"]; {
	case rowtype == "control" && key == "accountcode":
		return ""
	case rowtype == "doctotal" && key == "description":
		return fill(language.Text("gl_report_doc_total", lang), []string{row["linecount"]})
	case rowtype == "budgettotal" && key == "budgetname":
		return language.Text("gl_report_total", lang)
	case rowtype == "daytotal" && key == "description" && row["linecount"] != "":
		return fill(language.Text("gl_report_day_total_count", lang), []string{row["linecount"]})
	case key == "description" && glRowLabels[rowtype] != "":
		return language.Text(glRowLabels[rowtype], lang)
	}
	return value
}

// csvCell quotes every cell (RFC 4180) and neutralises formula text; an amount cell keeps a plain number such as
// "-100.00" as is (a spreadsheet must read it as a number) but anything else in it is guarded like text
func csvCell(value string, amount bool) string {
	if formulaStart.MatchString(value) && !(amount && isDecimal(value)) {
		value = "'" + value
	}
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

// WriteCSV streams the report as UTF-8 CSV with BOM and CRLF: header ("<group> <label>"), rows, then footer rows
func WriteCSV(w io.Writer, resp RunResponse, lang string) error {
	out := bufio.NewWriterSize(w, 64<<10)
	if _, err := out.WriteString("\uFEFF"); err != nil {
		return err
	}
	cells := make([]string, len(resp.Columns))
	for i, c := range resp.Columns {
		title := c.Label
		if c.Group != "" {
			title = c.Group + " " + c.Label
		}
		cells[i] = csvCell(title, false)
	}
	if err := writeLine(out, cells); err != nil {
		return err
	}
	for _, row := range resp.Rows {
		for i, c := range resp.Columns {
			cells[i] = csvCell(cellText(row, c.Key, i == 0, lang), isAmount(c))
		}
		if err := writeLine(out, cells); err != nil {
			return err
		}
	}
	for _, row := range resp.Footer {
		for i, c := range resp.Columns {
			value := row[c.Key]
			if i == 0 {
				value = row["label"]
			}
			cells[i] = csvCell(value, isAmount(c) && i > 0)
		}
		if err := writeLine(out, cells); err != nil {
			return err
		}
	}
	return out.Flush()
}

func isAmount(c ColumnView) bool { return c.Type == ColAmount || c.Type == ColQty }

func isDecimal(value string) bool {
	_, err := decimal.NewFromString(value)
	return err == nil
}

func writeLine(w *bufio.Writer, cells []string) error {
	_, err := w.WriteString(strings.Join(cells, ",") + "\r\n")
	return err
}

// ContentDisposition = attachment with an ASCII fallback name and the UTF-8 report name (RFC 6266 / 5987)
func ContentDisposition(asciiName, utf8Name string) string {
	return `attachment; filename="` + asciiName + `"; filename*=UTF-8''` + rfc5987(utf8Name)
}

// rfc5987 percent-encodes every UTF-8 byte except unreserved characters, so the header has no ; , or quotes
func rfc5987(value string) string {
	var b strings.Builder
	for _, c := range []byte(value) {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '.' || c == '_' || c == '~' {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}
