package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"smlcloudplatform/internal/authentication/models"
)

type AuthenticationPostgresRepository struct {
	db *sql.DB
}

func NewAuthenticationPostgresRepository(db *sql.DB) IAuthenticationMongoCacheRepository {
	return &AuthenticationPostgresRepository{db: db}
}

func (r *AuthenticationPostgresRepository) FindUser(ctx context.Context, username string) (*models.UserDoc, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	var (
		id           uuid.UUID
		uName        string
		passwordHash string
		email        sql.NullString
		phone        sql.NullString
		fullName     string
		isActive     bool
		createdAt    time.Time
		updatedAt    time.Time
	)

	query := `SELECT id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at
	          FROM users WHERE LOWER(username) = $1 LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&id, &uName, &passwordHash, &email, &phone, &fullName, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	doc := &models.UserDoc{
		GuidFixed: id.String(),
		UsernameField: models.UsernameField{
			Username: uName,
		},
		EmailField: models.EmailField{
			Email: email.String,
		},
		PhoneNumberField: models.PhoneNumberField{
			PhoneNumber: phone.String,
		},
		UserPassword: models.UserPassword{
			Password: passwordHash,
		},
		UserDetail: models.UserDetail{
			UID:  id.String(),
			Name: fullName,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		IsDeleted: !isActive,
	}

	// Synthesize primitive.ObjectID for backward compatibility with Mongo ObjectId calls
	objID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024x", id.ID()))
	doc.ID = objID

	return doc, nil
}

func (r *AuthenticationPostgresRepository) FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error) {
	var query string
	var arg interface{}

	switch strings.ToLower(fieldName) {
	case "uid", "guidfixed":
		query = `SELECT id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at
		          FROM users WHERE id::text = $1 LIMIT 1`
		arg = value
	case "username":
		return r.FindUser(ctx, value)
	case "email":
		query = `SELECT id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at
		          FROM users WHERE LOWER(email) = LOWER($1) LIMIT 1`
		arg = value
	case "phone", "phonenumber":
		query = `SELECT id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at
		          FROM users WHERE phone = $1 LIMIT 1`
		arg = value
	default:
		query = `SELECT u.id, u.username, u.password_hash, u.email, u.phone, u.full_name, u.is_active, u.created_at, u.updated_at
		          FROM users u
		          JOIN user_identities i ON u.id = i.user_id
		          WHERE i.provider = $1 AND i.identity_id = $2 LIMIT 1`
		var (
			id           uuid.UUID
			uName        string
			passwordHash string
			email        sql.NullString
			phone        sql.NullString
			fullName     string
			isActive     bool
			createdAt    time.Time
			updatedAt    time.Time
		)
		err := r.db.QueryRowContext(ctx, query, fieldName, value).Scan(
			&id, &uName, &passwordHash, &email, &phone, &fullName, &isActive, &createdAt, &updatedAt,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, mongo.ErrNoDocuments
			}
			return nil, err
		}
		doc := &models.UserDoc{
			GuidFixed: id.String(),
			UsernameField: models.UsernameField{
				Username: uName,
			},
			EmailField: models.EmailField{
				Email: email.String,
			},
			PhoneNumberField: models.PhoneNumberField{
				PhoneNumber: phone.String,
			},
			UserPassword: models.UserPassword{
				Password: passwordHash,
			},
			UserDetail: models.UserDetail{
				UID:  id.String(),
				Name: fullName,
			},
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			IsDeleted: !isActive,
		}
		objID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024x", id.ID()))
		doc.ID = objID
		return doc, nil
	}

	var (
		id           uuid.UUID
		uName        string
		passwordHash string
		email        sql.NullString
		phone        sql.NullString
		fullName     string
		isActive     bool
		createdAt    time.Time
		updatedAt    time.Time
	)
	err := r.db.QueryRowContext(ctx, query, arg).Scan(
		&id, &uName, &passwordHash, &email, &phone, &fullName, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	doc := &models.UserDoc{
		GuidFixed: id.String(),
		UsernameField: models.UsernameField{
			Username: uName,
		},
		EmailField: models.EmailField{
			Email: email.String,
		},
		PhoneNumberField: models.PhoneNumberField{
			PhoneNumber: phone.String,
		},
		UserPassword: models.UserPassword{
			Password: passwordHash,
		},
		UserDetail: models.UserDetail{
			UID:  id.String(),
			Name: fullName,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		IsDeleted: !isActive,
	}
	objID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024x", id.ID()))
	doc.ID = objID
	return doc, nil
}

func (r *AuthenticationPostgresRepository) FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error) {
	phone := strings.TrimSpace(phonenumber.PhoneNumber)
	var (
		id           uuid.UUID
		uName        string
		passwordHash string
		email        sql.NullString
		phoneVal     sql.NullString
		fullName     string
		isActive     bool
		createdAt    time.Time
		updatedAt    time.Time
	)

	query := `SELECT id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at
	          FROM users WHERE phone = $1 LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, phone).Scan(
		&id, &uName, &passwordHash, &email, &phoneVal, &fullName, &isActive, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	doc := &models.UserDoc{
		GuidFixed: id.String(),
		UsernameField: models.UsernameField{
			Username: uName,
		},
		EmailField: models.EmailField{
			Email: email.String,
		},
		PhoneNumberField: models.PhoneNumberField{
			CountryCode: phonenumber.CountryCode,
			PhoneNumber: phoneVal.String,
		},
		UserPassword: models.UserPassword{
			Password: passwordHash,
		},
		UserDetail: models.UserDetail{
			UID:  id.String(),
			Name: fullName,
		},
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		IsDeleted: !isActive,
	}
	objID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024x", id.ID()))
	doc.ID = objID
	return doc, nil
}

func (r *AuthenticationPostgresRepository) CreateUser(ctx context.Context, doc models.UserDoc) (primitive.ObjectID, error) {
	newID := uuid.New()
	if doc.GuidFixed != "" {
		if parsed, err := uuid.Parse(doc.GuidFixed); err == nil {
			newID = parsed
		}
	}

	query := `INSERT INTO users (id, username, password_hash, email, phone, full_name, is_active, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, true, now(), now())
	          ON CONFLICT (username) DO UPDATE
	          SET password_hash = EXCLUDED.password_hash,
	              full_name = EXCLUDED.full_name,
	              updated_at = now()
	          RETURNING id`
	var returnedID uuid.UUID
	err := r.db.QueryRowContext(ctx, query,
		newID,
		strings.ToLower(strings.TrimSpace(doc.Username)),
		doc.Password,
		doc.Email,
		doc.PhoneNumber,
		doc.Name,
	).Scan(&returnedID)
	if err != nil {
		return primitive.NilObjectID, err
	}

	objID, _ := primitive.ObjectIDFromHex(fmt.Sprintf("%024x", returnedID.ID()))
	return objID, nil
}

func (r *AuthenticationPostgresRepository) UpdateUser(ctx context.Context, username string, user models.UserDoc) error {
	query := `UPDATE users
	          SET password_hash = CASE WHEN $2 <> '' THEN $2 ELSE password_hash END,
	              full_name = CASE WHEN $3 <> '' THEN $3 ELSE full_name END,
	              email = CASE WHEN $4 <> '' THEN $4 ELSE email END,
	              phone = CASE WHEN $5 <> '' THEN $5 ELSE phone END,
	              updated_at = now()
	          WHERE LOWER(username) = LOWER($1)`
	_, err := r.db.ExecContext(ctx, query,
		strings.ToLower(strings.TrimSpace(username)),
		user.Password,
		user.Name,
		user.Email,
		user.PhoneNumber,
	)
	return err
}

func (r *AuthenticationPostgresRepository) DeleteUser(ctx context.Context, username string) error {
	query := `UPDATE users SET is_active = false, updated_at = now() WHERE LOWER(username) = LOWER($1)`
	_, err := r.db.ExecContext(ctx, query, strings.ToLower(strings.TrimSpace(username)))
	return err
}

func (r *AuthenticationPostgresRepository) FindByLineUserID(ctx context.Context, lineUserID string) (*models.UserDoc, error) {
	return r.FindByIdentity(ctx, "line", lineUserID)
}

func (r *AuthenticationPostgresRepository) FindUserByUID(ctx context.Context, userUID string) (*models.UserDoc, error) {
	return r.FindByIdentity(ctx, "uid", userUID)
}

func (r *AuthenticationPostgresRepository) UpdateUserByUID(ctx context.Context, userUID string, user models.UserDoc) error {
	query := `UPDATE users
	          SET password_hash = CASE WHEN $2 <> '' THEN $2 ELSE password_hash END,
	              full_name = CASE WHEN $3 <> '' THEN $3 ELSE full_name END,
	              email = CASE WHEN $4 <> '' THEN $4 ELSE email END,
	              phone = CASE WHEN $5 <> '' THEN $5 ELSE phone END,
	              updated_at = now()
	          WHERE id::text = $1`
	_, err := r.db.ExecContext(ctx, query,
		userUID,
		user.Password,
		user.Name,
		user.Email,
		user.PhoneNumber,
	)
	return err
}

func (r *AuthenticationPostgresRepository) FindGoogleIdentity(ctx context.Context, issuer string, subject string) (*models.GoogleIdentity, error) {
	return findPostgresGoogleIdentity(ctx, r.db, issuer, subject)
}

func (r *AuthenticationPostgresRepository) CreateAuthAudit(ctx context.Context, audit models.AuthAudit) error {
	return nil
}

func (r *AuthenticationPostgresRepository) CreateGoogleUserIdentity(ctx context.Context, user models.UserDoc, identity models.GoogleIdentity, audit models.AuthAudit) (models.UserDoc, error) {
	return r.createPostgresGoogleIdentity(ctx, user, identity, audit)
}

func (r *AuthenticationPostgresRepository) EnsureGoogleIdentityIndexes(ctx context.Context) error {
	return nil
}
