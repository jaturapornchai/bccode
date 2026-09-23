package whtcert

import (
	"strings"

	"github.com/shopspring/decimal"
)

var (
	thaiDigits = []string{"ศูนย์", "หนึ่ง", "สอง", "สาม", "สี่", "ห้า", "หก", "เจ็ด", "แปด", "เก้า"}
	thaiUnits  = []string{"", "สิบ", "ร้อย", "พัน", "หมื่น", "แสน"}
)

// BahtText - จำนวนเงินเป็นตัวอักษรไทย เช่น 1234.50 → "หนึ่งพันสองร้อยสามสิบสี่บาทห้าสิบสตางค์"
// กติกาเดียวกับ frontend/src/lib/thai-baht-text.ts (เอ็ด / ยี่สิบ / สิบ / ล้านซ้อน)
func BahtText(amount decimal.Decimal) string {
	fixed := amount.Abs().StringFixed(2)
	intPart, satang, _ := strings.Cut(fixed, ".")
	if intPart == "0" && satang == "00" {
		return "ศูนย์บาทถ้วน"
	}

	text := ""
	if intPart != "0" {
		var groups []string
		for rest := intPart; rest != ""; {
			take := min(6, len(rest))
			groups = append([]string{rest[len(rest)-take:]}, groups...)
			rest = rest[:len(rest)-take]
		}
		multiDigit := len(intPart) > 1 || intPart > "1"
		for i, group := range groups {
			groupText := thaiDigitGroup(group, i == len(groups)-1 && multiDigit)
			if groupText != "" {
				text += groupText + strings.Repeat("ล้าน", len(groups)-i-1)
			}
		}
		text += "บาท"
	}

	if satang == "00" {
		text += "ถ้วน"
	} else {
		text += thaiDigitGroup(satang, satang > "01") + "สตางค์"
	}
	if amount.Sign() < 0 {
		return "ลบ" + text
	}
	return text
}

// thaiDigitGroup - อ่านตัวเลขไม่เกิน 6 หลัก (หนึ่งรอบล้าน)
func thaiDigitGroup(digits string, unitOneIsEt bool) string {
	result := ""
	for i, ch := range digits {
		digit := int(ch - '0')
		pos := len(digits) - i - 1
		switch {
		case digit == 0:
		case pos == 0 && digit == 1 && unitOneIsEt:
			result += "เอ็ด"
		case pos == 1 && digit == 2:
			result += "ยี่สิบ"
		case pos == 1 && digit == 1:
			result += "สิบ"
		default:
			result += thaiDigits[digit] + thaiUnits[pos]
		}
	}
	return result
}
