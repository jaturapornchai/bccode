package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb"
)

// ErrNotFound is returned (with an empty, non-nil document) when a user or identity does not exist.
var ErrNotFound = errors.New("record not found")

// ErrUserExists is returned by CreateUser when the username is already taken.
var ErrUserExists = errors.New("username is exists")

type IAuthenticationRepository interface {
	FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error)
	FindUser(ctx context.Context, username string) (*models.UserDoc, error)
	FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error)
	FindByLineUserID(ctx context.Context, lineUserID string) (*models.UserDoc, error)
	CreateUser(ctx context.Context, doc models.UserDoc) (string, error)
	UpdateUser(ctx context.Context, username string, user models.UserDoc) error
	UpdateUserByUID(ctx context.Context, userUID string, user models.UserDoc) error
	SetLineIdentity(ctx context.Context, userUID string, lineUserID string, displayName string, pictureURL string) error
	DeleteUser(ctx context.Context, username string) error
	FindGoogleIdentity(ctx context.Context, issuer string, subject string) (*models.GoogleIdentity, error)
	FindUserByUID(ctx context.Context, userUID string) (*models.UserDoc, error)
	CreateGoogleUserIdentity(ctx context.Context, user models.UserDoc, identity models.GoogleIdentity, audit models.AuthAudit) (models.UserDoc, error)
}

type AuthenticationPostgresRepository struct {
	db *sql.DB
}

func NewAuthenticationPostgresRepository(db *sql.DB) IAuthenticationRepository {
	return &AuthenticationPostgresRepository{db: db}
}

const userColumns = `u.id::text, u.username, u.password_hash, COALESCE(u.email, ''), COALESCE(u.phone, ''), u.full_name, u.is_active, u.created_at, u.updated_at,
	COALESCE(l.identity_id, ''), COALESCE(l.extra->>'displayName', ''), COALESCE(l.extra->>'pictureUrl', '')`

const userFrom = ` FROM users u LEFT JOIN user_identities l ON l.user_id = u.id AND l.provider = 'line' `

// findOne scans one user. A missing user yields an empty document plus ErrNotFound,
// so callers that read fields before checking the error never dereference nil.
func (r *AuthenticationPostgresRepository) findOne(ctx context.Context, where string, args ...interface{}) (*models.UserDoc, error) {
	doc := &models.UserDoc{}
	var isActive bool
	err := r.db.QueryRowContext(ctx, `SELECT `+userColumns+userFrom+where+` LIMIT 1`, args...).Scan(
		&doc.UID, &doc.Username, &doc.Password, &doc.Email, &doc.PhoneNumber, &doc.Name, &isActive, &doc.CreatedAt, &doc.UpdatedAt,
		&doc.LineUserID, &doc.LineDisplayName, &doc.LinePictureURL,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &models.UserDoc{}, ErrNotFound
	}
	if err != nil {
		return &models.UserDoc{}, err
	}
	doc.GuidFixed = doc.UID
	doc.ID = doc.UID
	if !isActive {
		doc.DisabledAt = doc.UpdatedAt
	}
	return doc, nil
}

func (r *AuthenticationPostgresRepository) FindUser(ctx context.Context, username string) (*models.UserDoc, error) {
	return r.findOne(ctx, `WHERE LOWER(u.username) = $1`, strings.ToLower(strings.TrimSpace(username)))
}

func (r *AuthenticationPostgresRepository) FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error) {
	switch strings.ToLower(fieldName) {
	case "uid", "guidfixed":
		return r.FindUserByUID(ctx, value)
	case "username":
		return r.FindUser(ctx, value)
	case "email":
		return r.findOne(ctx, `WHERE LOWER(u.email) = LOWER($1) ORDER BY u.created_at`, strings.TrimSpace(value))
	case "phone", "phonenumber":
		return r.findOne(ctx, `WHERE u.phone = $1`, strings.TrimSpace(value))
	default:
		return r.findOne(ctx, `WHERE u.id = (SELECT user_id FROM user_identities WHERE provider = $1 AND identity_id = $2)`,
			strings.ToLower(fieldName), value)
	}
}

func (r *AuthenticationPostgresRepository) FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error) {
	doc, err := r.findOne(ctx, `WHERE u.phone = $1`, strings.TrimSpace(phonenumber.PhoneNumber))
	if err == nil {
		doc.CountryCode = phonenumber.CountryCode
	}
	return doc, err
}

func (r *AuthenticationPostgresRepository) FindByLineUserID(ctx context.Context, lineUserID string) (*models.UserDoc, error) {
	return r.FindByIdentity(ctx, "line", lineUserID)
}

func (r *AuthenticationPostgresRepository) FindUserByUID(ctx context.Context, userUID string) (*models.UserDoc, error) {
	if _, err := uuid.Parse(strings.TrimSpace(userUID)); err != nil {
		return &models.UserDoc{}, ErrNotFound
	}
	return r.findOne(ctx, `WHERE u.id = $1`, strings.TrimSpace(userUID))
}

// CreateUser inserts a new user and never overwrites an existing account.
func (r *AuthenticationPostgresRepository) CreateUser(ctx context.Context, doc models.UserDoc) (string, error) {
	newID := uuid.NewString()
	for _, candidate := range []string{doc.UID, doc.GuidFixed} {
		if parsed, err := uuid.Parse(strings.TrimSpace(candidate)); err == nil {
			newID = parsed.String()
			break
		}
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO users (id, username, password_hash, email, phone, full_name, is_active)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), $6, true)`,
		newID, strings.ToLower(strings.TrimSpace(doc.Username)), doc.Password, strings.TrimSpace(doc.Email), strings.TrimSpace(doc.PhoneNumber), doc.Name)
	if centraldb.IsUniqueViolation(err) {
		return "", ErrUserExists
	}
	if err != nil {
		return "", err
	}
	return newID, nil
}

// updateUser writes profile fields; blank values keep what is stored. A non-zero
// DisabledAt deactivates the account.
func (r *AuthenticationPostgresRepository) updateUser(ctx context.Context, where string, key string, user models.UserDoc) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users u
		SET password_hash = CASE WHEN $2 <> '' THEN $2 ELSE password_hash END,
		    full_name = CASE WHEN $3 <> '' THEN $3 ELSE full_name END,
		    email = CASE WHEN $4 <> '' THEN $4 ELSE email END,
		    phone = CASE WHEN $5 <> '' THEN $5 ELSE phone END,
		    is_active = $6,
		    updated_at = now()
		`+where, key, user.Password, user.Name, user.Email, user.PhoneNumber, user.DisabledAt.IsZero())
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *AuthenticationPostgresRepository) UpdateUser(ctx context.Context, username string, user models.UserDoc) error {
	return r.updateUser(ctx, `WHERE LOWER(u.username) = $1`, strings.ToLower(strings.TrimSpace(username)), user)
}

func (r *AuthenticationPostgresRepository) UpdateUserByUID(ctx context.Context, userUID string, user models.UserDoc) error {
	if _, err := uuid.Parse(strings.TrimSpace(userUID)); err != nil {
		return ErrNotFound
	}
	return r.updateUser(ctx, `WHERE u.id = $1`, strings.TrimSpace(userUID), user)
}

// SetLineIdentity links a LINE account to the user (replacing any previous link);
// an empty lineUserID unlinks it.
func (r *AuthenticationPostgresRepository) SetLineIdentity(ctx context.Context, userUID string, lineUserID string, displayName string, pictureURL string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_identities WHERE user_id = $1 AND provider = 'line'`, userUID); err != nil {
		return err
	}
	if lineUserID = strings.TrimSpace(lineUserID); lineUserID != "" {
		extra, err := json.Marshal(map[string]string{"displayName": displayName, "pictureUrl": pictureURL})
		if err != nil {
			return err
		}
		_, err = tx.ExecContext(ctx, `INSERT INTO user_identities (user_id, provider, identity_id, extra) VALUES ($1, 'line', $2, $3)`,
			userUID, lineUserID, string(extra))
		if centraldb.IsUniqueViolation(err) {
			return ErrUserExists
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

// DeleteUser removes the account; memberships and identities cascade.
func (r *AuthenticationPostgresRepository) DeleteUser(ctx context.Context, username string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE LOWER(username) = $1`, strings.ToLower(strings.TrimSpace(username)))
	return err
}

func (r *AuthenticationPostgresRepository) FindGoogleIdentity(ctx context.Context, issuer string, subject string) (*models.GoogleIdentity, error) {
	return findPostgresGoogleIdentity(ctx, r.db, issuer, subject)
}

func (r *AuthenticationPostgresRepository) CreateGoogleUserIdentity(ctx context.Context, user models.UserDoc, identity models.GoogleIdentity, audit models.AuthAudit) (models.UserDoc, error) {
	return r.createPostgresGoogleIdentity(ctx, user, identity, audit)
}
