package utils

import (
	"strings"
)

func NormalizeUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.ToLower(username)
	return username
}

func NormalizePhonenumber(phoneNumber string) string {
	phoneNumber = strings.TrimSpace(phoneNumber)
	phoneNumber = strings.TrimPrefix(phoneNumber, "0")
	return phoneNumber
}

func NormalizeEmail(email string) string {
	email = strings.TrimSpace(email)
	return email
}

func NormalizeName(username string) string {
	username = strings.TrimSpace(username)
	return username
}
