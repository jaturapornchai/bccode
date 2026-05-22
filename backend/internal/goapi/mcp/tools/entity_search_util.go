package tools

import (
	"regexp"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

// ==================== Smart Entity Keyword Tokenization ====================
//
// ปัญหา: ผู้ใช้พิมพ์ "ร้านวัฒนา" แต่ในฐานข้อมูลเก็บว่า "บริษัท วัฒนา ก่อสร้าง จำกัด"
// การใช้ regex ตรง ๆ จะหาไม่เจอ
//
// วิธีแก้:
// 1. Strip Thai business prefixes (ร้าน, บริษัท, หจก, บจก, ห้างหุ้นส่วน, etc.)
// 2. Split on whitespace + Thai word boundaries
// 3. Build OR of regex for each token (ต้องเจออย่างน้อย 1 token)

// thaiBusinessPrefixes คำนำหน้าชื่อธุรกิจไทย (ต้อง strip ออก)
var thaiBusinessPrefixes = []string{
	"บริษัท จำกัด (มหาชน)",
	"บริษัทจำกัด(มหาชน)",
	"บริษัท จำกัด",
	"บริษัทจำกัด",
	"บจก.",
	"บจก",
	"ห้างหุ้นส่วนจำกัด",
	"ห้างหุ้นส่วน",
	"หจก.",
	"หจก",
	"ร้าน",
	"ร.ร.",
	"บ.",
	"company limited",
	"co., ltd.",
	"co.,ltd.",
	"co.ltd.",
	"co ltd",
	"ltd.",
	"ltd",
	"limited",
	"inc.",
	"inc",
	"corp.",
	"corp",
}

// mongoRegexEscape escape ตัวอักษรพิเศษใน regex
var mongoRegexEscape = regexp.MustCompile(`[.*+?()\[\]{}|^$\\]`)

// escapeRegex escape special regex characters for MongoDB regex search
func escapeRegex(s string) string {
	return mongoRegexEscape.ReplaceAllStringFunc(s, func(m string) string {
		return `\` + m
	})
}

// TokenizeEntityKeyword ถอดคำนำหน้าธุรกิจ + แยกคำ
// Returns list of meaningful tokens (min length 2)
func TokenizeEntityKeyword(keyword string) []string {
	if keyword == "" {
		return nil
	}
	lower := strings.ToLower(strings.TrimSpace(keyword))

	// Strip business prefixes (case-insensitive)
	for _, prefix := range thaiBusinessPrefixes {
		p := strings.ToLower(prefix)
		// Strip from start
		if strings.HasPrefix(lower, p) {
			lower = strings.TrimSpace(lower[len(p):])
		}
		// Strip from end
		if strings.HasSuffix(lower, p) {
			lower = strings.TrimSpace(lower[:len(lower)-len(p)])
		}
		// Strip from middle (contains)
		lower = strings.ReplaceAll(lower, p, " ")
	}

	// Split on whitespace + common separators
	lower = strings.NewReplacer(
		",", " ",
		"/", " ",
		"(", " ",
		")", " ",
		"-", " ",
	).Replace(lower)

	fields := strings.Fields(lower)

	// Also split continuous Thai+ASCII "ร้านวัฒนา" → try to split Thai prefix
	// Add both the full keyword and each token
	result := make([]string, 0, len(fields)+1)
	seen := make(map[string]bool)
	for _, f := range fields {
		if len([]rune(f)) >= 2 && !seen[f] {
			result = append(result, f)
			seen[f] = true
		}
	}

	// If stripping removed everything, fall back to original
	if len(result) == 0 {
		orig := strings.TrimSpace(keyword)
		// Try stripping Thai prefix from original (for cases where no space)
		for _, prefix := range []string{"ร้าน", "บริษัท", "หจก", "บจก"} {
			if strings.HasPrefix(orig, prefix) {
				rest := strings.TrimSpace(orig[len(prefix):])
				if len([]rune(rest)) >= 2 {
					result = append(result, rest)
					break
				}
			}
		}
		if len(result) == 0 && len([]rune(orig)) >= 2 {
			result = append(result, orig)
		}
	}

	// Also try to split "ร้านวัฒนา" → "วัฒนา" if no space but starts with prefix
	for _, prefix := range []string{"ร้าน", "บริษัท", "หจก", "บจก"} {
		for _, tok := range fields {
			if strings.HasPrefix(tok, prefix) {
				rest := strings.TrimSpace(tok[len(prefix):])
				if len([]rune(rest)) >= 2 && !seen[rest] {
					result = append(result, rest)
					seen[rest] = true
				}
			}
		}
	}

	return result
}

// isThaiRune ตรวจสอบว่า rune เป็นอักษรไทย (U+0E00..U+0E7F)
func isThaiRune(r rune) bool {
	return r >= 0x0E00 && r <= 0x0E7F
}

// isMostlyThai true ถ้า token มีอักษรไทยเป็นหลัก (>= 80%)
func isMostlyThai(s string) bool {
	total := 0
	thai := 0
	for _, r := range s {
		if r == ' ' || r == '\t' {
			continue
		}
		total++
		if isThaiRune(r) {
			thai++
		}
	}
	if total == 0 {
		return false
	}
	return thai*100/total >= 80
}

// buildFlexibleThaiRegex สร้าง regex ที่ยอมให้มี space ระหว่างตัวอักษรไทยได้
// เช่น "อรุณโฮม" → "อ\s*ร\s*ุ\s*ณ\s*โ\s*ฮ\s*ม" จะจับได้ทั้ง "อรุณโฮม" และ "อรุณ โฮม"
// ใช้เฉพาะ token ภาษาไทยล้วนที่ยาว >= 4 runes (สั้นกว่านั้นเสี่ยง false positive มาก)
func buildFlexibleThaiRegex(token string) string {
	runes := []rune(token)
	if len(runes) < 4 || !isMostlyThai(token) {
		return escapeRegex(token)
	}
	var sb strings.Builder
	for i, r := range runes {
		if i > 0 {
			sb.WriteString(`\s*`)
		}
		sb.WriteString(escapeRegex(string(r)))
	}
	return sb.String()
}

// BuildEntityKeywordFilter สร้าง mongo filter สำหรับค้นหา entity (debtor/creditor/customer)
// ใช้ตัวแปร fieldNames เช่น ["code", "names.name", "tax_id"]
// Logic: tokenize keyword → any token matches any field (OR of OR)
// สำหรับ token ภาษาไทยล้วน >= 4 runes จะใช้ regex ยืดหยุ่น (space-insensitive)
func BuildEntityKeywordFilter(keyword string, fieldNames []string) bson.M {
	tokens := TokenizeEntityKeyword(keyword)
	if len(tokens) == 0 {
		tokens = []string{strings.TrimSpace(keyword)}
	}

	var orClauses []bson.M
	for _, tok := range tokens {
		pattern := buildFlexibleThaiRegex(tok)
		for _, field := range fieldNames {
			orClauses = append(orClauses, bson.M{
				field: bson.M{"$regex": pattern, "$options": "i"},
			})
		}
	}

	return bson.M{"$or": orClauses}
}
