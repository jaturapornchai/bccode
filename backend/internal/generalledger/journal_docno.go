package generalledger

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Running document numbers the Champ way (champ-lib BOLEDBCOMM/BRunningBttn.cpp OnRun):
// the format is read left to right — YY = Buddhist year (2 digits), MM = month, DD = day,
// a run of @ = the journal book code, # = one running digit; Champ cuts the format at the
// first # and counts every # as a digit, and so do we. The next number is MAX(existing)+1
// within the same prefix. Unlike Champ the date comes from the voucher date (not today),
// so vouchers keyed late still number into their own month.

const (
	DefaultJournalDocFormat = "@YYMM-####"
	docFormatMaxRunes       = 20
)

var (
	errDocFormatInvalid = fieldError("journal_book_docformat_invalid", "docformat", "รูปแบบเลขที่เอกสารต้องมีเครื่องหมาย # อย่างน้อย 1 ตัวสำหรับเลขรัน และยาวไม่เกิน 20 ตัวอักษร")
	errDocNoTooLong     = fieldError("journal_docno_auto_too_long", "docno", "เลขที่เอกสารที่สร้างจากรูปแบบของสมุดรายวันยาวเกิน 30 ตัวอักษร กรุณาแก้รูปแบบเลขที่เอกสารของสมุด")
)

func validateDocFormat(format string) error {
	if format == "" {
		return nil
	}
	if !strings.Contains(format, "#") || utf8.RuneCountInString(format) > docFormatMaxRunes {
		return errDocFormatInvalid
	}
	return nil
}

// docNoPrefix expands format for a book and voucher date (YYYY-MM-DD) into the fixed prefix
// and the number of running digits.
func docNoPrefix(format, book, date string) (string, int, error) {
	if format == "" {
		format = DefaultJournalDocFormat
	}
	if err := validateDocFormat(format); err != nil {
		return "", 0, err
	}
	if !validDate(date) {
		return "", 0, fieldError("journal_date_invalid", "date", "วันที่เอกสารไม่ถูกต้อง กรุณาเลือกวันที่ให้ครบ วัน เดือน ปี")
	}
	year, _ := strconv.Atoi(date[:4])
	values := map[string]string{"YY": fmt.Sprintf("%02d", (year+543)%100), "MM": date[5:7], "DD": date[8:10]}
	upper := []rune(strings.ToUpper(format))
	var prefix strings.Builder
	digits := strings.Count(format, "#")
	for i := 0; i < len(upper); i++ {
		switch r := upper[i]; {
		case r == '#':
			return prefix.String(), digits, nil
		case r == '@':
			for i+1 < len(upper) && upper[i+1] == '@' {
				i++
			}
			prefix.WriteString(book)
		case i+1 < len(upper) && values[string(upper[i:i+2])] != "":
			prefix.WriteString(values[string(upper[i:i+2])])
			i++
		default:
			prefix.WriteRune(r)
		}
	}
	return prefix.String(), digits, nil
}

// nextJournalDocNo returns the next free number for the book's format and the voucher date.
// Callers run inside Execute, which already holds the company lock (lockCompany), so two
// users saving at once cannot draw the same number. Deleted drafts keep their number, so a
// number is never reused.
func nextJournalDocNo(ctx context.Context, tx *sql.Tx, company string, book Master, date string) (string, error) {
	prefix, digits, err := docNoPrefix(book.DocFormat, book.Code, date)
	if err != nil {
		return "", err
	}
	var last int64
	err = tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(substr(code, char_length($2)+1)::bigint), 0) FROM gl_records
		WHERE company=$1 AND kind='journals' AND left(code, char_length($2))=$2 AND substr(code, char_length($2)+1) ~ '^[0-9]{1,18}$'`, company, prefix).Scan(&last)
	if err != nil {
		return "", err
	}
	docno := prefix + fmt.Sprintf("%0*d", digits, last+1)
	if utf8.RuneCountInString(docno) > DocNoMaxRunes {
		return "", errDocNoTooLong
	}
	return docno, nil
}

// PreviewJournalDocNo shows the number the next new voucher of a book would get. It is only a
// preview: the real number is drawn when the voucher is saved, inside the saving transaction.
func (s *PostgresStore) PreviewJournalDocNo(ctx context.Context, scope Scope, bookCode, date string) (string, error) {
	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return "", err
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return "", err
	}
	defer tx.Rollback()
	book, found, err := loadJournalBook(ctx, tx, scope.Company, bookCode)
	if err != nil {
		return "", err
	}
	if !found {
		return "", journalBookNotFound(bookCode)
	}
	return nextJournalDocNo(ctx, tx, scope.Company, book, date)
}
