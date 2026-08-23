package models

import (
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	msvalidator "smlcloudplatform/pkg/validator"
)

func TestUsercodeNormalizationAndValidation(t *testing.T) {
	if got := NormalizeUsercode("  Jead.Admin_01  "); got != "jead.admin_01" {
		t.Fatalf("NormalizeUsercode() = %q, want %q", got, "jead.admin_01")
	}

	tests := []struct {
		name     string
		usercode string
		valid    bool
	}{
		{name: "minimum", usercode: "abc", valid: true},
		{name: "allowed punctuation", usercode: "jead.admin-01_test", valid: true},
		{name: "normalizes case and surrounding space", usercode: "  ABC  ", valid: true},
		{name: "maximum", usercode: strings.Repeat("a", 64), valid: true},
		{name: "too short", usercode: "ab", valid: false},
		{name: "too long", usercode: strings.Repeat("a", 65), valid: false},
		{name: "email is not usercode", usercode: "jead@example.com", valid: false},
		{name: "internal space", usercode: "jead admin", valid: false},
		{name: "thai", usercode: "ลุงจืด", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsValidUsercode(test.usercode); got != test.valid {
				t.Fatalf("IsValidUsercode(%q) = %v, want %v", test.usercode, got, test.valid)
			}
		})
	}
}

func TestGoogleOnlyUserOmitsUnsetPasswordCredentialsFromMongo(t *testing.T) {
	raw, err := bson.Marshal(UserDoc{EmailField: EmailField{Email: "user@example.com"}})
	if err != nil {
		t.Fatal(err)
	}
	var document bson.M
	if err := bson.Unmarshal(raw, &document); err != nil {
		t.Fatal(err)
	}
	if _, exists := document["username"]; exists {
		t.Fatal("unset username must be omitted so the sparse unique index remains usable")
	}
	if _, exists := document["password"]; exists {
		t.Fatal("Google-only user must not persist an empty password field")
	}
}

func TestPasswordRequestLengthValidation(t *testing.T) {
	validator := msvalidator.NewCustomValidator()

	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{name: "minimum", password: strings.Repeat("a", 15), valid: true},
		{name: "maximum", password: strings.Repeat("a", 64), valid: true},
		{name: "too short", password: strings.Repeat("a", 14), valid: false},
		{name: "too long", password: strings.Repeat("a", 65), valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := UserLoginRequest{
				UsernameField: UsernameField{Username: "jead_admin"},
				UserPassword:  UserPassword{Password: test.password},
			}
			err := validator.Validate(request)
			if test.valid && err != nil {
				t.Fatalf("valid password rejected: %v", err)
			}
			if !test.valid && err == nil {
				t.Fatal("invalid password accepted")
			}
		})
	}
}

func TestKnownCompromisedPassword(t *testing.T) {
	if !IsKnownCompromisedPassword("  Password123456  ") {
		t.Fatal("known compromised password must be rejected case-insensitively")
	}
	if IsKnownCompromisedPassword("unique-passphrase-2026") {
		t.Fatal("unlisted passphrase must not be classified as compromised")
	}
}
