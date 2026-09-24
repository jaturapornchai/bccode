package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/authentication/models"

	"github.com/google/uuid"
)

const googleIssuer = "https://accounts.google.com"

// This one legacy Google account was created without an identity link. Its owner
// authorized repair; pinning its immutable ID avoids general email-based linking.
const legacyDeveloperUID = "da5ef3fe-7aa9-4e4e-8011-0a0dec3c417a"
const legacyDeveloperEmail = "jaturapornchai@gmail.com"

type googleQuery interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type googleIdentityExtra struct {
	Issuer        string            `json:"issuer"`
	VerifiedEmail string            `json:"verifiedEmail"`
	Active        *bool             `json:"active,omitempty"`
	RevokedAt     *time.Time        `json:"revokedAt,omitempty"`
	Audit         *models.AuthAudit `json:"linkAudit,omitempty"`
}

func findPostgresGoogleIdentity(ctx context.Context, db googleQuery, issuer, subject string) (*models.GoogleIdentity, error) {
	if (issuer != googleIssuer && issuer != "accounts.google.com") || strings.TrimSpace(subject) == "" {
		return nil, errors.New("invalid Google identity")
	}
	identity := &models.GoogleIdentity{Issuer: googleIssuer, Subject: subject, IsActive: true}
	var extra []byte
	err := db.QueryRowContext(ctx, `SELECT id::text,user_id::text,extra,created_at FROM user_identities WHERE provider='google' AND identity_id=$1`, subject).Scan(&identity.IdentityUID, &identity.UserUID, &extra, &identity.LinkedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var metadata googleIdentityExtra
	if len(extra) > 0 && json.Unmarshal(extra, &metadata) != nil {
		return nil, errors.New("invalid Google identity metadata")
	}
	if metadata.Issuer != "" && metadata.Issuer != googleIssuer {
		return nil, errors.New("Google issuer mismatch")
	}
	identity.VerifiedEmail, identity.RevokedAt = metadata.VerifiedEmail, metadata.RevokedAt
	if metadata.Active != nil {
		identity.IsActive = *metadata.Active
	}
	return identity, nil
}

func (r *AuthenticationPostgresRepository) createPostgresGoogleIdentity(ctx context.Context, input models.UserDoc, identity models.GoogleIdentity, audit models.AuthAudit) (models.UserDoc, error) {
	var empty models.UserDoc
	email := strings.ToLower(strings.TrimSpace(identity.VerifiedEmail))
	if identity.Issuer != googleIssuer || strings.TrimSpace(identity.Subject) == "" || email == "" || !strings.EqualFold(strings.TrimSpace(input.Email), email) || !identity.IsActive || identity.RevokedAt != nil {
		return empty, errors.New("invalid verified Google identity")
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return empty, err
	}
	defer tx.Rollback()
	// A concurrent first login must reuse the committed identity, not create an orphan user.
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, "google:"+identity.Subject); err != nil {
		return empty, err
	}
	linked, err := findPostgresGoogleIdentity(ctx, tx, identity.Issuer, identity.Subject)
	if err == nil {
		if !linked.IsActive || linked.RevokedAt != nil {
			return empty, errors.New("Google identity is inactive")
		}
		return postgresGoogleUser(ctx, tx, linked.UserUID)
	}
	if !errors.Is(err, ErrNotFound) {
		return empty, err
	}
	uid, precreated, err := googleUserForFirstLink(ctx, tx, email, input.Name)
	if err != nil {
		return empty, err
	}
	audit.UserUID = uid
	if precreated != nil {
		// Who and when: this identity row (created_at) + the audit say which Google account
		// claimed which admin-created login account.
		audit.Action = "GOOGLE_IDENTITY_LINK_PRECREATED"
		if audit.Metadata == nil {
			audit.Metadata = map[string]interface{}{}
		}
		audit.Metadata["matched_by"] = precreated.matchedBy
		audit.Metadata["precreated_username"] = precreated.username
		audit.Metadata["verified_email"] = email
		audit.Metadata["linked_by"] = "google:" + identity.Subject
	}
	active := true
	extra, err := json.Marshal(googleIdentityExtra{Issuer: googleIssuer, VerifiedEmail: email, Active: &active, Audit: &audit})
	if err != nil {
		return empty, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO user_identities(id,user_id,provider,identity_id,extra) VALUES($1,$2,'google',$3,$4)`, uuid.NewString(), uid, identity.Subject, string(extra)); err != nil {
		return empty, err
	}
	user, err := postgresGoogleUser(ctx, tx, uid)
	if err != nil {
		return empty, err
	}
	if err = tx.Commit(); err != nil {
		return empty, err
	}
	return user, nil
}

// precreatedLogin is the admin-created login account a first Google login attached to.
type precreatedLogin struct {
	uid, username string
	matchedBy     string // "email" or "username"
}

// googleUserForFirstLink picks the account a Google identity seen for the first time signs in
// to: the pinned legacy developer account, else the one account an admin pre-created for this
// verified email (findPrecreatedLogin), else a new "google_<uuid>" account.
func googleUserForFirstLink(ctx context.Context, tx *sql.Tx, email, name string) (string, *precreatedLogin, error) {
	uid, err := legacyDeveloperForFirstLink(ctx, tx, email)
	if err != nil || uid != "" {
		return uid, nil, err
	}
	precreated, err := findPrecreatedLogin(ctx, tx, email)
	if err != nil {
		return "", nil, err
	}
	if precreated != nil {
		// Fill only what the admin left blank; never overwrite what they entered.
		_, err = tx.ExecContext(ctx, `UPDATE users SET
			email = CASE WHEN COALESCE(TRIM(email), '') = '' THEN $2 ELSE email END,
			full_name = CASE WHEN TRIM(full_name) = '' THEN $3 ELSE full_name END,
			updated_at = now()
			WHERE id = $1`, precreated.uid, email, strings.TrimSpace(name))
		return precreated.uid, precreated, err
	}
	uid = uuid.NewString()
	_, err = tx.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,email,full_name,is_active) VALUES($1,$2,'',$3,$4,true)`, uid, "google_"+strings.ReplaceAll(uid, "-", ""), email, strings.TrimSpace(name))
	return uid, nil, err
}

// precreatedEligible is an account an admin created (Settings › Login Accounts) that nobody has
// signed in to yet: it has a usercode, belongs to a Holding, is active, and has no password and
// no linked identity of any provider. Its email or usercode must equal the Google-verified email
// ($1, lower-case) exactly, ignoring case. Legacy rows without a usercode are never matched.
const precreatedEligible = `u.is_active AND u.password_hash = '' AND u.username <> ''
	AND (LOWER(TRIM(COALESCE(u.email, ''))) = $1 OR LOWER(TRIM(u.username)) = $1)
	AND NOT EXISTS (SELECT 1 FROM user_identities i WHERE i.user_id = u.id)
	AND EXISTS (SELECT 1 FROM holding_members m WHERE m.user_id = u.id)`

// ErrGooglePrecreatedAmbiguous: more than one pre-created account matches the Google-verified
// email. The login is refused (the message asks the user to contact their admin) instead of
// creating a fresh account: that account would bind the Google identity for good, so the
// account the real admin prepared could never be linked — anyone able to add a member with
// that email in some Holding could block (or hijack) another Holding's invitation.
var ErrGooglePrecreatedAmbiguous = errors.New("more than one pre-created login account matches this Google email")

// findPrecreatedLogin returns the single pre-created account for email, nil when there is
// none, or ErrGooglePrecreatedAmbiguous (wrapped with the account ids for the log) when more than one
// match (ambiguous: never guess which person this is, and never merge accounts).
// A claimed account (password or any identity) is never taken over: its owner already proved
// who they are, so an email match alone must not hand it to someone else.
func findPrecreatedLogin(ctx context.Context, tx *sql.Tx, email string) (*precreatedLogin, error) {
	rows, err := tx.QueryContext(ctx, `SELECT u.id::text FROM users u WHERE `+precreatedEligible+` ORDER BY u.id FOR UPDATE OF u`, email)
	if err != nil {
		return nil, err
	}
	locked := []string{}
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		locked = append(locked, id)
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if len(locked) > 1 {
		return nil, fmt.Errorf("%w: accounts %s", ErrGooglePrecreatedAmbiguous, strings.Join(locked, ","))
	}
	if len(locked) == 0 {
		return nil, nil
	}
	// Re-check in a fresh statement: a statement that waited for the row lock still evaluates
	// its sub-queries on the snapshot it started with, so an identity linked or an email changed
	// while we waited would be missed.
	link := &precreatedLogin{}
	var byEmail bool
	err = tx.QueryRowContext(ctx, `SELECT u.id::text, u.username, LOWER(TRIM(COALESCE(u.email, ''))) = $1
		FROM users u WHERE u.id = $2::uuid AND `+precreatedEligible, email, locked[0]).Scan(&link.uid, &link.username, &byEmail)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	link.matchedBy = "username"
	if byEmail {
		link.matchedBy = "email"
	}
	return link, nil
}

// legacyDeveloperForFirstLink repairs the one pinned legacy account (see legacyDeveloperUID);
// "" when email is not its email or the account no longer exists.
func legacyDeveloperForFirstLink(ctx context.Context, tx *sql.Tx, email string) (string, error) {
	if email == legacyDeveloperEmail {
		var storedEmail, username, password string
		var active bool
		err := tx.QueryRowContext(ctx, `SELECT COALESCE(email,''),username,password_hash,is_active FROM users WHERE id=$1 FOR UPDATE`, legacyDeveloperUID).Scan(&storedEmail, &username, &password, &active)
		if err == nil {
			var linked bool
			if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_identities WHERE user_id=$1)`, legacyDeveloperUID).Scan(&linked); err != nil {
				return "", err
			}
			if !active || username != "" || password != "" || !strings.EqualFold(strings.TrimSpace(storedEmail), email) || linked {
				return "", errors.New("legacy Google account cannot be linked automatically")
			}
			if _, err = tx.ExecContext(ctx, `UPDATE users SET username=$2,updated_at=now() WHERE id=$1`, legacyDeveloperUID, "google_"+strings.ReplaceAll(legacyDeveloperUID, "-", "")); err != nil {
				return "", err
			}
			return legacyDeveloperUID, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	return "", nil
}

func postgresGoogleUser(ctx context.Context, db googleQuery, uid string) (models.UserDoc, error) {
	var user models.UserDoc
	err := db.QueryRowContext(ctx, `SELECT id::text,username,COALESCE(email,''),full_name,created_at,updated_at FROM users WHERE id=$1 AND is_active=true`, uid).Scan(&user.UID, &user.Username, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	user.GuidFixed = user.UID
	return user, err
}
