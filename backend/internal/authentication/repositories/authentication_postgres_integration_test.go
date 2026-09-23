//go:build integration

package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb/centraldbtest"
)

func TestAuthenticationPostgresRepositoryAgainstCentralSchema(t *testing.T) {
	db := centraldbtest.New(t)
	ctx := context.Background()
	repo := NewAuthenticationPostgresRepository(db)

	missing, err := repo.FindUser(ctx, "nobody")
	if !errors.Is(err, ErrNotFound) || missing == nil || missing.Username != "" {
		t.Fatalf("missing user = %+v %v", missing, err)
	}

	user := models.UserDoc{}
	user.Username = " Somchai@Example.com "
	user.Email = "somchai@example.com"
	user.Password = "hash-1"
	user.Name = "สมชาย ใจดี"
	uid, err := repo.CreateUser(ctx, user)
	if err != nil || uid == "" {
		t.Fatalf("CreateUser = %q %v", uid, err)
	}
	user.Password = "attacker-hash"
	if _, err := repo.CreateUser(ctx, user); !errors.Is(err, ErrUserExists) {
		t.Fatalf("duplicate CreateUser error = %v", err)
	}

	found, err := repo.FindUser(ctx, "SOMCHAI@example.com")
	if err != nil || found.UID != uid || found.Password != "hash-1" || !found.DisabledAt.IsZero() {
		t.Fatalf("FindUser = %+v %v (password must not be overwritten)", found, err)
	}
	if byEmail, err := repo.FindByIdentity(ctx, "email", "SOMCHAI@EXAMPLE.COM"); err != nil || byEmail.UID != uid {
		t.Fatalf("FindByIdentity(email) = %+v %v", byEmail, err)
	}
	if _, err := repo.FindUserByUID(ctx, "not-a-uuid"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("invalid uid error = %v", err)
	}

	if err := repo.SetLineIdentity(ctx, uid, "U123", "Somchai", "https://profile.line-scdn.net/a"); err != nil {
		t.Fatal(err)
	}
	byLine, err := repo.FindByLineUserID(ctx, "U123")
	if err != nil || byLine.UID != uid || byLine.LineDisplayName != "Somchai" {
		t.Fatalf("FindByLineUserID = %+v %v", byLine, err)
	}
	if err := repo.SetLineIdentity(ctx, uid, "", "", ""); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindByLineUserID(ctx, "U123"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unlinked LINE still resolves: %v", err)
	}

	found.DisabledAt = time.Now()
	if err := repo.UpdateUserByUID(ctx, uid, *found); err != nil {
		t.Fatal(err)
	}
	disabled, err := repo.FindUserByUID(ctx, uid)
	if err != nil || disabled.DisabledAt.IsZero() {
		t.Fatalf("disabled user = %+v %v", disabled, err)
	}

	centraldbtest.Exec(t, db, `INSERT INTO holdings (code, name) VALUES ('h', 'Holding')`)
	centraldbtest.Exec(t, db, `INSERT INTO holding_members (holding_code, user_id, role) VALUES ('h', $1, 'USER')`, uid)
	if err := repo.DeleteUser(ctx, "somchai@example.com"); err != nil {
		t.Fatal(err)
	}
	var members int
	if err := db.QueryRow(`SELECT COUNT(*) FROM holding_members WHERE user_id = $1`, uid).Scan(&members); err != nil || members != 0 {
		t.Fatalf("memberships after delete = %d %v", members, err)
	}
}
