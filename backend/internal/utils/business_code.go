package utils

import "strings"

// NormalizeBusinessCode keeps business keys stable across tables and API payloads.
func NormalizeBusinessCode(value string) string {
	return strings.ToUpper(strings.Join(strings.Fields(value), ""))
}
