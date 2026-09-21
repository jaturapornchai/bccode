package mcptoken

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTokenCredential(t *testing.T) {
	id, raw, hash, err := generate("H_test", "mcp")
	if err != nil {
		t.Fatal(err)
	}
	holding, parsed, err := parse(raw, "mcp")
	if err != nil || holding != "H_test" || parsed != id {
		t.Fatal("token round trip failed")
	}
	sum := sha256.Sum256([]byte(raw))
	if string(hash) != string(sum[:]) || strings.Contains(string(hash), raw) {
		t.Fatal("invalid stored credential")
	}
	_, second, _, _ := generate("H_test", "mcp")
	if raw == second {
		t.Fatal("reused secret")
	}
	for _, bad := range []string{"", raw + "x", "bcaimcp_Li4u." + id + ".bad", strings.Replace(raw, id, "../", 1), strings.Repeat("a", 201)} {
		if _, _, e := parse(bad, "mcp"); e == nil {
			t.Fatal("accepted malformed token")
		}
	}
}
func TestExpiryAndMode(t *testing.T) {
	now := time.Now().UTC()
	good := createInput{CompanyCodes: []string{"C"}, Kind: "mcp", Name: "test", Mode: "readonly", ExpiresAt: now.Add(time.Hour)}
	if !validateInput(good, now) {
		t.Fatal("rejected valid input")
	}
	for _, expiry := range []time.Time{now, now.Add(-time.Second), now.AddDate(1, 0, 1)} {
		in := good
		in.ExpiresAt = expiry
		if validateInput(in, now) {
			t.Fatal("accepted invalid expiry")
		}
	}
	good.Mode = "admin"
	if validateInput(good, now) {
		t.Fatal("accepted invalid mode")
	}
}
func TestAuthenticationFailsClosedWithoutDatabase(t *testing.T) {
	_, raw, _, _ := generate("H", "mcp")
	_, err := authenticateAudience(context.Background(), raw, "mcp", func(string) (*sql.DB, error) { return nil, errors.New("unavailable") }, time.Now())
	if !errors.Is(err, ErrDenied) {
		t.Fatal("database failure did not deny")
	}
}
