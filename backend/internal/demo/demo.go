// Package demo holds the configuration of the public Demo account: a single
// username-registered user that anyone can sign in as from the login screen
// (button "ทดลองใช้ระบบ") to explore realistic sample data. It works in every
// environment (local + public server) and is enabled by env only.
package demo

import (
	"os"
	"strings"
)

const defaultUsername = "demo"

// Enabled reports whether BCAI_DEMO_LOGIN_ENABLED=true.
func Enabled() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv("BCAI_DEMO_LOGIN_ENABLED")), "true")
}

// Username is the demo account's usercode (BCAI_DEMO_USERNAME, default "demo").
func Username() string {
	if value := strings.ToLower(strings.TrimSpace(os.Getenv("BCAI_DEMO_USERNAME"))); value != "" {
		return value
	}
	return defaultUsername
}

// IsDemoUser is true when demo login is enabled and username is the demo account.
// The demo account has no Google identity, so organization creation lets it through
// (docs/login.md "บัญชี Demo").
func IsDemoUser(username string) bool {
	return Enabled() && strings.EqualFold(strings.TrimSpace(username), Username())
}
