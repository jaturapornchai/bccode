package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
	uid, err := googleUserForFirstLink(ctx, tx, email, input.Name)
	if err != nil {
		return empty, err
	}
	audit.UserUID = uid
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

func googleUserForFirstLink(ctx context.Context, tx *sql.Tx, email, name string) (string, error) {
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
	uid := uuid.NewString()
	_, err := tx.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,email,full_name,is_active) VALUES($1,$2,'',$3,$4,true)`, uid, "google_"+strings.ReplaceAll(uid, "-", ""), email, strings.TrimSpace(name))
	return uid, err
}

func postgresGoogleUser(ctx context.Context, db googleQuery, uid string) (models.UserDoc, error) {
	var user models.UserDoc
	err := db.QueryRowContext(ctx, `SELECT id::text,username,COALESCE(email,''),full_name,created_at,updated_at FROM users WHERE id=$1 AND is_active=true`, uid).Scan(&user.UID, &user.Username, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	user.GuidFixed = user.UID
	return user, err
}
