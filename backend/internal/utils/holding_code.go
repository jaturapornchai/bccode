package utils

import (
	"errors"
	"regexp"
	"strings"
)

var holdingCodePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,29}$`)

func NormalizeHoldingCode(value string) (string, error) {
	holdingCode := strings.ToLower(strings.TrimSpace(value))
	if holdingCode == "" {
		return "", nil
	}
	if !holdingCodePattern.MatchString(holdingCode) {
		return "", errors.New("holding_code invalid")
	}
	return holdingCode, nil
}
