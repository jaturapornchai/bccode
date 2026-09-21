//go:build integration

package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"smlcloudplatform/internal/authentication/models"
)

func googleTestRepository(t *testing.T) *AuthenticationPostgresRepository {
	t.Helper()
	dsn := os.Getenv("MCP_TOKEN_TEST_DSN")
	if dsn == "" {
		t.Skip("MCP_TOKEN_TEST_DSN required")
	}
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	namespace := fmt.Sprintf("google_identity_%d", time.Now().UnixNano())
	if _, err = admin.Exec("CREATE SCHEMA " + namespace); err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	query := u.Query()
	query.Set("search_path", namespace)
	u.RawQuery = query.Encode()
	db, err := sql.Open("postgres", u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close(); admin.Exec("DROP SCHEMA " + namespace + " CASCADE"); admin.Close() })
	_, err = db.Exec(`CREATE TABLE users(id uuid PRIMARY KEY,username text UNIQUE NOT NULL,password_hash text NOT NULL,email text,phone text,full_name text NOT NULL DEFAULT '',is_active boolean NOT NULL DEFAULT true,created_at timestamptz NOT NULL DEFAULT now(),updated_at timestamptz NOT NULL DEFAULT now());
 CREATE TABLE user_identities(id uuid PRIMARY KEY,user_id uuid NOT NULL REFERENCES users(id),provider text NOT NULL,identity_id text NOT NULL,extra jsonb DEFAULT '{}',created_at timestamptz NOT NULL DEFAULT now(),UNIQUE(provider,identity_id));`)
	if err != nil {
		t.Fatal(err)
	}
	return &AuthenticationPostgresRepository{db: db}
}

func googleTestCreate(r *AuthenticationPostgresRepository, subject, email string) (models.UserDoc, error) {
	user := models.UserDoc{GuidFixed: uuid.NewString()}
	user.UID = uuid.NewString()
	user.Email = email
	user.Name = "Google user"
	identity := models.GoogleIdentity{Issuer: googleIssuer, Subject: subject, VerifiedEmail: email, IsActive: true, UserUID: user.UID}
	return r.CreateGoogleUserIdentity(context.Background(), user, identity, models.AuthAudit{Action: "GOOGLE_IDENTITY_LINK", Outcome: "SUCCESS", UserUID: user.UID})
}

func TestPostgresGoogleCanonicalIdentity(t *testing.T) {
	r := googleTestRepository(t)
	first, err := googleTestCreate(r, "first-subject", "first@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if first.UID == "" || first.UID != first.GuidFixed || !strings.HasPrefix(first.Username, "google_") {
		t.Fatal("noncanonical Google user")
	}
	linked, err := r.FindGoogleIdentity(context.Background(), googleIssuer, "first-subject")
	if err != nil || linked.UserUID != first.UID || linked.VerifiedEmail != "first@example.com" {
		t.Fatalf("missing persisted link: %v", err)
	}
	// Subsequent service logins load this path before creating the session.
	stored, err := r.FindUserByUID(context.Background(), linked.UserUID)
	if err != nil || stored.UID != first.UID || stored.GuidFixed != first.UID || stored.IsDeleted {
		t.Fatalf("linked login did not resolve canonical session user: %v", err)
	}
	repeat, err := googleTestCreate(r, "first-subject", "changed@example.com")
	if err != nil || repeat.UID != first.UID {
		t.Fatalf("repeat login changed identity: %v", err)
	}
	other, err := googleTestCreate(r, "second-subject", "first@example.com")
	if err != nil || other.UID == first.UID {
		t.Fatal("email linked unrelated subjects")
	}
	var count int
	if err = r.db.QueryRow(`SELECT count(*) FROM user_identities WHERE user_id::text=$1 AND extra->'linkAudit'->>'useruid'=$1`, first.UID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("audit did not bind canonical user: %v", err)
	}
	if _, err = r.FindGoogleIdentity(context.Background(), "https://evil.example", "first-subject"); err == nil {
		t.Fatal("wrong issuer accepted")
	}
	if _, err = r.db.Exec(`UPDATE users SET is_active=false WHERE id=$1`, first.UID); err != nil {
		t.Fatal(err)
	}
	if _, err = googleTestCreate(r, "first-subject", "first@example.com"); err == nil {
		t.Fatal("inactive user accepted")
	}
}

func TestPostgresGoogleConcurrentFirstLogin(t *testing.T) {
	r := googleTestRepository(t)
	var wg sync.WaitGroup
	results := make(chan models.UserDoc, 8)
	failures := make(chan error, 8)
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, e := googleTestCreate(r, "concurrent", "concurrent@example.com")
			if e != nil {
				failures <- e
			} else {
				results <- u
			}
		}()
	}
	wg.Wait()
	close(results)
	close(failures)
	for err := range failures {
		t.Fatal(err)
	}
	uid := ""
	for u := range results {
		if uid != "" && uid != u.UID {
			t.Fatal("concurrent duplicate identity")
		}
		uid = u.UID
	}
	for _, table := range []string{"users", "user_identities"} {
		var count int
		if err := r.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); err != nil || count != 1 {
			t.Fatalf("unexpected %s count: %d (%v)", table, count, err)
		}
	}
}

func TestPostgresGooglePinnedLegacyRepair(t *testing.T) {
	r := googleTestRepository(t)
	passwordUID := uuid.NewString()
	_, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email) VALUES($1,'jaturapornchai','password-hash',$3),($2,'','',$3)`, passwordUID, legacyDeveloperUID, legacyDeveloperEmail)
	if err != nil {
		t.Fatal(err)
	}
	u, err := googleTestCreate(r, "developer-subject", legacyDeveloperEmail)
	if err != nil || u.UID != legacyDeveloperUID {
		t.Fatalf("legacy UID not preserved: %v", err)
	}
	repeat, err := googleTestCreate(r, "developer-subject", legacyDeveloperEmail)
	if err != nil || repeat.UID != legacyDeveloperUID {
		t.Fatal("repaired account not stable")
	}
	if _, err = googleTestCreate(r, "different-developer-subject", legacyDeveloperEmail); err == nil {
		t.Fatal("different subject took linked legacy account")
	}
	var password string
	if err = r.db.QueryRow(`SELECT password_hash FROM users WHERE id=$1`, passwordUID).Scan(&password); err != nil || password != "password-hash" {
		t.Fatal("password account changed")
	}
}

func TestPostgresGoogleLegacyRepairBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, username, password, email string
		inactive, linked                bool
	}{
		{name: "email mismatch", email: "other@example.com"},
		{name: "username present", username: "real-user", email: legacyDeveloperEmail},
		{name: "password present", password: "hash", email: legacyDeveloperEmail},
		{name: "inactive", email: legacyDeveloperEmail, inactive: true},
		{name: "linked identity", email: legacyDeveloperEmail, linked: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := googleTestRepository(t)
			if _, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email,is_active) VALUES($1,$2,$3,$4,$5)`, legacyDeveloperUID, tc.username, tc.password, tc.email, !tc.inactive); err != nil {
				t.Fatal(err)
			}
			if tc.linked {
				if _, err := r.db.Exec(`INSERT INTO user_identities(id,user_id,provider,identity_id) VALUES($1,$2,'line','existing')`, uuid.NewString(), legacyDeveloperUID); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := googleTestCreate(r, "repair", legacyDeveloperEmail); err == nil {
				t.Fatal("unsafe legacy repair accepted")
			}
		})
	}
}

func TestPostgresGoogleLinkFailureRollsBack(t *testing.T) {
	r := googleTestRepository(t)
	if _, err := r.db.Exec(`ALTER TABLE user_identities ADD CONSTRAINT reject_link CHECK(identity_id <> 'fail')`); err != nil {
		t.Fatal(err)
	}
	if _, err := googleTestCreate(r, "fail", "new@example.com"); err == nil {
		t.Fatal("link failure ignored")
	}
	var count int
	if err := r.db.QueryRow(`SELECT count(*) FROM users`).Scan(&count); err != nil || count != 0 {
		t.Fatal("orphan user persisted")
	}
	if _, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email) VALUES($1,'','',$2)`, legacyDeveloperUID, legacyDeveloperEmail); err != nil {
		t.Fatal(err)
	}
	if _, err := googleTestCreate(r, "fail", legacyDeveloperEmail); err == nil {
		t.Fatal("legacy link failure ignored")
	}
	var name string
	if err := r.db.QueryRow(`SELECT username FROM users WHERE id=$1`, legacyDeveloperUID).Scan(&name); err != nil || name != "" {
		t.Fatal("legacy repair did not roll back")
	}
}

func TestPostgresGoogleDoesNotLinkGeneralLegacyEmail(t *testing.T) {
	r := googleTestRepository(t)
	oldUID := uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email) VALUES($1,'','','legacy@example.com')`, oldUID); err != nil {
		t.Fatal(err)
	}
	user, err := googleTestCreate(r, "other-legacy-subject", "legacy@example.com")
	if err != nil || user.UID == oldUID {
		t.Fatal("general email-based legacy linking")
	}
}

func TestPostgresGoogleRejectsRevokedIdentity(t *testing.T) {
	r := googleTestRepository(t)
	if _, err := googleTestCreate(r, "revoked", "revoked@example.com"); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`UPDATE user_identities SET extra=extra || '{"active":false,"revokedAt":"2026-09-20T00:00:00Z"}'::jsonb`); err != nil {
		t.Fatal(err)
	}
	identity, err := r.FindGoogleIdentity(context.Background(), googleIssuer, "revoked")
	if err != nil || identity.IsActive || identity.RevokedAt == nil {
		t.Fatal("revocation metadata ignored")
	}
	if _, err := googleTestCreate(r, "revoked", "revoked@example.com"); err == nil {
		t.Fatal("revoked identity accepted")
	}
}
