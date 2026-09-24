//go:build integration

package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"smlcloudplatform/internal/authentication/models"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
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
 CREATE TABLE user_identities(id uuid PRIMARY KEY,user_id uuid NOT NULL REFERENCES users(id),provider text NOT NULL,identity_id text NOT NULL,extra jsonb DEFAULT '{}',created_at timestamptz NOT NULL DEFAULT now(),UNIQUE(provider,identity_id));
 CREATE TABLE holding_members(holding_code text NOT NULL,user_id uuid NOT NULL REFERENCES users(id),PRIMARY KEY(holding_code,user_id));`)
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

// googleTestMember adds uid to a Holding, as Settings › Login Accounts does for the accounts it creates.
func googleTestMember(t *testing.T, r *AuthenticationPostgresRepository, uid string) {
	t.Helper()
	if _, err := r.db.Exec(`INSERT INTO holding_members(holding_code,user_id) VALUES('rungrueng',$1)`, uid); err != nil {
		t.Fatal(err)
	}
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

// A person an admin created in Settings › Login Accounts signs in with Google for the first
// time: the verified email (any case) attaches the Google login to that account instead of
// creating a second, empty one (save-audit finding: pre-created users could never log in).
func TestPostgresGoogleLinksPrecreatedAccountByEmail(t *testing.T) {
	r := googleTestRepository(t)
	precreated := uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email,full_name) VALUES($1,'somchai01','','Somchai@Company.co.th','สมชาย ใจดี')`, precreated); err != nil {
		t.Fatal(err)
	}
	googleTestMember(t, r, precreated)
	u, err := googleTestCreate(r, "somchai-subject", "somchai@company.co.th")
	if err != nil || u.UID != precreated || u.Username != "somchai01" || u.Name != "สมชาย ใจดี" {
		t.Fatalf("pre-created account not linked: %+v (%v)", u, err)
	}
	var action, matchedBy, precreatedUsername, linkedBy, verified string
	err = r.db.QueryRow(`SELECT extra->'linkAudit'->>'action', extra->'linkAudit'->'metadata'->>'matched_by',
		extra->'linkAudit'->'metadata'->>'precreated_username', extra->'linkAudit'->'metadata'->>'linked_by',
		extra->'linkAudit'->'metadata'->>'verified_email'
		FROM user_identities WHERE user_id=$1 AND provider='google'`, precreated).Scan(&action, &matchedBy, &precreatedUsername, &linkedBy, &verified)
	if err != nil || action != "GOOGLE_IDENTITY_LINK_PRECREATED" || matchedBy != "email" || precreatedUsername != "somchai01" ||
		linkedBy != "google:somchai-subject" || verified != "somchai@company.co.th" {
		t.Fatalf("link audit = %q %q %q %q %q (%v)", action, matchedBy, precreatedUsername, linkedBy, verified, err)
	}
	// The account is now claimed: another Google account with the same email gets its own user.
	other, err := googleTestCreate(r, "other-subject", "somchai@company.co.th")
	if err != nil || other.UID == precreated || !strings.HasPrefix(other.Username, "google_") {
		t.Fatalf("claimed account taken by a second Google subject: %+v (%v)", other, err)
	}
}

// The admin typed the email as the usercode and left the email blank: link by usercode and
// fill only the blanks.
func TestPostgresGoogleLinksPrecreatedAccountByUsername(t *testing.T) {
	r := googleTestRepository(t)
	precreated := uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email,full_name) VALUES($1,'malee@company.co.th','',NULL,'')`, precreated); err != nil {
		t.Fatal(err)
	}
	googleTestMember(t, r, precreated)
	u, err := googleTestCreate(r, "malee-subject", "MALEE@company.co.th")
	if err != nil || u.UID != precreated {
		t.Fatalf("pre-created account not linked by usercode: %+v (%v)", u, err)
	}
	var email, name, matchedBy string
	if err = r.db.QueryRow(`SELECT COALESCE(u.email,''), u.full_name, i.extra->'linkAudit'->'metadata'->>'matched_by'
		FROM users u JOIN user_identities i ON i.user_id=u.id WHERE u.id=$1`, precreated).Scan(&email, &name, &matchedBy); err != nil {
		t.Fatal(err)
	}
	if email != "malee@company.co.th" || name != "Google user" || matchedBy != "username" {
		t.Fatalf("blanks not filled: email=%q name=%q matched_by=%q", email, name, matchedBy)
	}
}

// An email match alone must never hand over an account whose owner already proved who they
// are, nor pick between two people, nor match a look-alike address.
func TestPostgresGoogleNeverTakesOverOrMergesAccounts(t *testing.T) {
	for _, tc := range []struct {
		name  string
		setup string // $1, $2 = two fresh ids
	}{
		{name: "has password", setup: `INSERT INTO users(id,username,password_hash,email) VALUES($1,'somchai01','hash','somchai@company.co.th')`},
		{name: "has another identity", setup: `INSERT INTO users(id,username,password_hash,email) VALUES($1,'somchai01','','somchai@company.co.th'); INSERT INTO user_identities(id,user_id,provider,identity_id) VALUES($2,$1,'line','line-user')`},
		{name: "inactive", setup: `INSERT INTO users(id,username,password_hash,email,is_active) VALUES($1,'somchai01','','somchai@company.co.th',false)`},
		{name: "not in any holding", setup: `INSERT INTO users(id,username,password_hash,email) VALUES($1,'somchai01','','somchai@company.co.th')`},
		{name: "look-alike email", setup: `INSERT INTO users(id,username,password_hash,email) VALUES($1,'somchai01','','somchai@company.co.th.example'),($2,'xsomchai@company.co.th','','')`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := googleTestRepository(t)
			first, second := uuid.NewString(), uuid.NewString()
			tx, err := r.db.Begin()
			if err != nil {
				t.Fatal(err)
			}
			for _, statement := range strings.Split(tc.setup, "; ") {
				args := []interface{}{first}
				if strings.Contains(statement, "$2") {
					args = append(args, second)
				}
				if _, err = tx.Exec(statement, args...); err != nil {
					tx.Rollback()
					t.Fatalf("%s: %v", statement, err)
				}
			}
			// Every account is a Holding member, as the admin screen makes them, except in "not in any holding".
			for _, uid := range []string{first, second} {
				if tc.name == "not in any holding" {
					break
				}
				if _, err = tx.Exec(`INSERT INTO holding_members(holding_code,user_id) SELECT 'rungrueng', id FROM users WHERE id=$1`, uid); err != nil {
					tx.Rollback()
					t.Fatal(err)
				}
			}
			if err = tx.Commit(); err != nil {
				t.Fatal(err)
			}
			u, err := googleTestCreate(r, "google-subject", "somchai@company.co.th")
			if err != nil || u.UID == first || u.UID == second || !strings.HasPrefix(u.Username, "google_") {
				t.Fatalf("existing account taken over: %+v (%v)", u, err)
			}
			var linked int
			if err = r.db.QueryRow(`SELECT count(*) FROM user_identities WHERE provider='google' AND user_id IN ($1,$2)`, first, second).Scan(&linked); err != nil || linked != 0 {
				t.Fatalf("google identity attached to an existing account: %d (%v)", linked, err)
			}
		})
	}
}

// Two pre-created accounts match the email (for example one added by another Holding's admin):
// the login is refused instead of creating a google_ account, which would bind the Google
// identity for good and block the real invitation (adversarial review 2026-09-24).
func TestPostgresGoogleRefusesAmbiguousPrecreatedAccounts(t *testing.T) {
	r := googleTestRepository(t)
	first, second := uuid.NewString(), uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email) VALUES($1,'somchai01','','somchai@company.co.th'),($2,'somchai@company.co.th','','')`, first, second); err != nil {
		t.Fatal(err)
	}
	if _, err := r.db.Exec(`INSERT INTO holding_members(holding_code,user_id) VALUES('rungrueng',$1),('other',$2)`, first, second); err != nil {
		t.Fatal(err)
	}
	_, err := googleTestCreate(r, "google-subject-ambiguous", "somchai@company.co.th")
	if !errors.Is(err, ErrGooglePrecreatedAmbiguous) || !strings.Contains(err.Error(), first) || !strings.Contains(err.Error(), second) {
		t.Fatalf("ambiguous match must be refused with both account ids: %v", err)
	}
	var users, identities int
	if err = r.db.QueryRow(`SELECT (SELECT count(*) FROM users WHERE username LIKE 'google%'), (SELECT count(*) FROM user_identities WHERE provider='google')`).Scan(&users, &identities); err != nil || users != 0 || identities != 0 {
		t.Fatalf("refused login left a google account (%d) or identity (%d): %v", users, identities, err)
	}
	// After the admin removes the stray account, the same Google login links the prepared one.
	if _, err = r.db.Exec(`DELETE FROM holding_members WHERE user_id=$1`, second); err != nil {
		t.Fatal(err)
	}
	if _, err = r.db.Exec(`DELETE FROM users WHERE id=$1`, second); err != nil {
		t.Fatal(err)
	}
	u, err := googleTestCreate(r, "google-subject-ambiguous", "somchai@company.co.th")
	if err != nil || u.UID != first {
		t.Fatalf("after cleanup the prepared account must link: %+v (%v)", u, err)
	}
}

// Two different Google accounts with the same verified email sign in at the same moment: at
// most one of them gets the pre-created account.
func TestPostgresGoogleConcurrentPrecreatedLink(t *testing.T) {
	r := googleTestRepository(t)
	precreated := uuid.NewString()
	if _, err := r.db.Exec(`INSERT INTO users(id,username,password_hash,email) VALUES($1,'somchai01','','somchai@company.co.th')`, precreated); err != nil {
		t.Fatal(err)
	}
	googleTestMember(t, r, precreated)
	var wg sync.WaitGroup
	uids := make(chan string, 6)
	for i := range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := googleTestCreate(r, fmt.Sprintf("subject-%d", i), "somchai@company.co.th")
			if err != nil {
				t.Error(err)
				return
			}
			uids <- u.UID
		}()
	}
	wg.Wait()
	close(uids)
	got := 0
	for uid := range uids {
		if uid == precreated {
			got++
		}
	}
	var identities int
	if err := r.db.QueryRow(`SELECT count(*) FROM user_identities WHERE user_id=$1`, precreated).Scan(&identities); err != nil {
		t.Fatal(err)
	}
	if got != 1 || identities != 1 {
		t.Fatalf("pre-created account linked %d times (%d identities), want exactly 1", got, identities)
	}
}
