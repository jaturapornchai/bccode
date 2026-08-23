package utils

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestPasswordHashArgon2IDRoundTrip(t *testing.T) {
	password := "รหัสผ่านที่ยาวและปลอดภัยมาก-2026"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if !CheckHashPassword(password, hash) {
		t.Fatal("correct password must verify")
	}
	if CheckHashPassword(password+"x", hash) {
		t.Fatal("wrong password must not verify")
	}
}

func TestPasswordHashStillVerifiesLegacyBcrypt(t *testing.T) {
	legacy, err := bcrypt.GenerateFromPassword([]byte("legacy-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("legacy hash: %v", err)
	}
	if !CheckHashPassword("legacy-password", string(legacy)) {
		t.Fatal("legacy bcrypt password must verify during transition")
	}
}
