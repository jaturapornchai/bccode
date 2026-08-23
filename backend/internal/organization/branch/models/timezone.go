package models

import (
	"errors"
	"strings"
	"time"
	_ "time/tzdata"
)

var (
	ErrBranchTimezoneRequired = errors.New("timezone is required")
	ErrBranchTimezoneInvalid  = errors.New("timezone must be a valid IANA timezone")
)

func NormalizeIANATimezone(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", ErrBranchTimezoneRequired
	}
	if strings.EqualFold(value, "Local") {
		return "", ErrBranchTimezoneInvalid
	}
	if _, err := time.LoadLocation(value); err != nil {
		return "", ErrBranchTimezoneInvalid
	}
	return value, nil
}
