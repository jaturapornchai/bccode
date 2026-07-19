package utils

import "strings"

// NormalizeBusinessCode keeps business keys stable across MongoDB and projections.
func NormalizeBusinessCode(value string) string {
	return strings.ToUpper(strings.Join(strings.Fields(value), ""))
}
