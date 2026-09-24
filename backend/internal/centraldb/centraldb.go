// Package centraldb owns the central control database (bcai_projection): identities,
// Holdings, Holding memberships, companies, branches and Holding-level master data.
package centraldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/lib/pq"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/mcptoken"
)

var (
	schemaMu    sync.Mutex
	schemaReady = map[*sql.DB]bool{}
)

// Open returns the pooled central database with the schema below ensured once per pool.
func Open() (*sql.DB, error) {
	db, err := mypg.PgSqlFastConnect(mcptoken.ControlDatabase)
	if err != nil {
		return nil, fmt.Errorf("connect central database: %w", err)
	}
	if db == nil {
		return nil, fmt.Errorf("connect central database: no connection")
	}
	if err := ensureOnce(db); err != nil {
		return nil, err
	}
	return db, nil
}

func ensureOnce(db *sql.DB) error {
	schemaMu.Lock()
	defer schemaMu.Unlock()
	if schemaReady[db] {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := EnsureSchema(ctx, db); err != nil {
		return err
	}
	schemaReady[db] = true
	return nil
}

// EnsureSchema creates the central tables and columns the backend reads and writes.
// Forward-only: new columns are added with IF NOT EXISTS, nothing is migrated.
func EnsureSchema(ctx context.Context, db *sql.DB) error {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("ensure central schema: %w", err)
	}
	defer conn.Close()
	// Serialize concurrent first starts of several API processes on one session.
	if _, err := conn.ExecContext(ctx, `SELECT pg_advisory_lock(hashtextextended('bcai_central_schema', 0))`); err != nil {
		return fmt.Errorf("ensure central schema: %w", err)
	}
	defer conn.ExecContext(context.Background(), `SELECT pg_advisory_unlock(hashtextextended('bcai_central_schema', 0))`)
	for _, statement := range schemaStatements {
		if _, err := conn.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("ensure central schema: %w", err)
		}
	}
	return nil
}

var schemaStatements = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		username TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL DEFAULT '',
		email TEXT,
		phone TEXT,
		full_name TEXT NOT NULL DEFAULT '',
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`,
	`CREATE TABLE IF NOT EXISTS user_identities (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		provider TEXT NOT NULL,
		identity_id TEXT NOT NULL,
		extra JSONB DEFAULT '{}'::jsonb,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		UNIQUE (provider, identity_id)
	)`,
	`CREATE TABLE IF NOT EXISTS holdings (
		code TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		tax_id TEXT DEFAULT '',
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`,
	`ALTER TABLE holdings ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT ''`,
	// Holding profile/settings (names, logo, language, currency, timezone ...) edited as one document.
	`ALTER TABLE holdings ADD COLUMN IF NOT EXISTS profile JSONB NOT NULL DEFAULT '{}'::jsonb`,
	`ALTER TABLE holdings ADD COLUMN IF NOT EXISTS updated_by TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE holdings ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`,
	`CREATE TABLE IF NOT EXISTS companies (
		holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
		code TEXT NOT NULL,
		name TEXT NOT NULL,
		tax_id TEXT DEFAULT '',
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		PRIMARY KEY (holding_code, code)
	)`,
	`ALTER TABLE companies ADD COLUMN IF NOT EXISTS names JSONB NOT NULL DEFAULT '[]'::jsonb`,
	`ALTER TABLE companies ADD COLUMN IF NOT EXISTS logo_uri TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE companies ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE companies ADD COLUMN IF NOT EXISTS updated_by TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE companies ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`,
	`CREATE TABLE IF NOT EXISTS branches (
		holding_code TEXT NOT NULL,
		company_code TEXT NOT NULL,
		code TEXT NOT NULL,
		name TEXT NOT NULL,
		is_headquarters BOOLEAN NOT NULL DEFAULT false,
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		PRIMARY KEY (holding_code, company_code, code),
		FOREIGN KEY (holding_code, company_code) REFERENCES companies(holding_code, code) ON DELETE CASCADE
	)`,
	`ALTER TABLE branches ADD COLUMN IF NOT EXISTS names JSONB NOT NULL DEFAULT '[]'::jsonb`,
	// Branch locale/tax/document settings are edited as one form (value-only configuration).
	`ALTER TABLE branches ADD COLUMN IF NOT EXISTS settings JSONB NOT NULL DEFAULT '{}'::jsonb`,
	`ALTER TABLE branches ADD COLUMN IF NOT EXISTS created_by TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE branches ADD COLUMN IF NOT EXISTS updated_by TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE branches ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now()`,
	`CREATE TABLE IF NOT EXISTS holding_members (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role TEXT NOT NULL DEFAULT 'USER',
		permission_sets JSONB NOT NULL DEFAULT '[]'::jsonb,
		access_scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		UNIQUE (holding_code, user_id)
	)`,
	`ALTER TABLE holding_members ADD COLUMN IF NOT EXISTS is_favorite BOOLEAN NOT NULL DEFAULT false`,
	`ALTER TABLE holding_members ADD COLUMN IF NOT EXISTS last_accessed_at TIMESTAMPTZ`,
	// Login-account screen fields are per membership: a person may hold another position,
	// department and picture in each business group, and one group's admin must not change
	// what another group shows. access_expiry_date: usable through the END of that date in the
	// Holding's timezone (AccessExpiryInstant → authmodels.AccessEndsAt); NULL = no expiry.
	`ALTER TABLE holding_members ADD COLUMN IF NOT EXISTS position TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE holding_members ADD COLUMN IF NOT EXISTS department TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE holding_members ADD COLUMN IF NOT EXISTS avatar TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE holding_members ADD COLUMN IF NOT EXISTS avatar_thumb TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE holding_members ADD COLUMN IF NOT EXISTS access_expiry_date DATE`,
	`CREATE TABLE IF NOT EXISTS role_permissions (
		holding_code TEXT NOT NULL REFERENCES holdings(code) ON DELETE CASCADE,
		role_code TEXT NOT NULL,
		permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
		PRIMARY KEY (holding_code, role_code)
	)`,
	`ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS id TEXT`,
	`ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS names JSONB DEFAULT '[]'::jsonb`,
	`ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true`,
	`ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS created_at TIMESTAMPTZ DEFAULT now()`,
	`ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ DEFAULT now()`,
	// Optimistic lock of a permission set (API field __v): every write adds 1, an update must
	// send the version it loaded or it is refused as changed meanwhile (409).
	`ALTER TABLE role_permissions ADD COLUMN IF NOT EXISTS version BIGINT NOT NULL DEFAULT 0`,
	`CREATE TABLE IF NOT EXISTS business_types (
		id TEXT PRIMARY KEY,
		holding_code TEXT NOT NULL,
		code TEXT NOT NULL,
		names JSONB NOT NULL DEFAULT '[]'::jsonb,
		is_default BOOLEAN NOT NULL DEFAULT false,
		is_active BOOLEAN NOT NULL DEFAULT true,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		UNIQUE (holding_code, code)
	)`,
	`CREATE TABLE IF NOT EXISTS employees (
		id TEXT PRIMARY KEY,
		holding_code TEXT NOT NULL,
		code TEXT NOT NULL,
		name TEXT NOT NULL,
		email TEXT NOT NULL DEFAULT '',
		roles JSONB NOT NULL DEFAULT '[]'::jsonb,
		is_enabled BOOLEAN NOT NULL DEFAULT true,
		is_use_pos BOOLEAN NOT NULL DEFAULT true,
		pin_code TEXT NOT NULL DEFAULT '',
		access_scopes JSONB NOT NULL DEFAULT '[]'::jsonb,
		branches JSONB NOT NULL DEFAULT '[]'::jsonb,
		contact JSONB NOT NULL DEFAULT '{}'::jsonb,
		profile_picture TEXT NOT NULL DEFAULT '',
		profile_picture_thumb TEXT NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		UNIQUE (holding_code, code)
	)`,
	`ALTER TABLE employees ADD COLUMN IF NOT EXISTS branches JSONB NOT NULL DEFAULT '[]'::jsonb`,
	// Append-only trail of organization structure changes (status changes carry a reason).
	`CREATE TABLE IF NOT EXISTS organization_audits (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		holding_code TEXT NOT NULL,
		action TEXT NOT NULL,
		target_type TEXT NOT NULL,
		target_code TEXT NOT NULL,
		company_code TEXT NOT NULL DEFAULT '',
		actor_uid TEXT NOT NULL DEFAULT '',
		reason TEXT NOT NULL DEFAULT '',
		before_state JSONB,
		after_state JSONB,
		occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`,
	`CREATE INDEX IF NOT EXISTS idx_organization_audits_target ON organization_audits (holding_code, target_type, target_code)`,
	`CREATE TABLE IF NOT EXISTS shop_user_access_logs (
		id BIGSERIAL PRIMARY KEY,
		holding_code TEXT NOT NULL,
		business_code TEXT NOT NULL DEFAULT '',
		username TEXT NOT NULL DEFAULT '',
		ip TEXT NOT NULL DEFAULT '',
		last_accessed_at TIMESTAMPTZ NOT NULL DEFAULT now()
	)`,
	// ตารางที่สร้างก่อนมีคอลัมน์นี้ (prod 2026-09) — CREATE IF NOT EXISTS ไม่เติมคอลัมน์ให้
	`ALTER TABLE shop_user_access_logs ADD COLUMN IF NOT EXISTS business_code TEXT NOT NULL DEFAULT ''`,
}

// IsUniqueViolation reports a PostgreSQL unique_violation (SQLSTATE 23505).
func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

// HoldingTimezoneSQL reads a Holding's timezone setting; h is the holdings row alias.
const HoldingTimezoneSQL = `COALESCE(h.profile->'settings'->>'timezone', '')`

// HoldingLocation is the Holding's timezone (authmodels.HoldingLocation — one rule for every package).
func HoldingLocation(timezone string) *time.Location {
	return authmodels.HoldingLocation(timezone)
}

// AccessExpiryInstant is the moment a membership's access ends: the expiry date is usable
// through its end ("ใช้งานได้ถึงสิ้นวันที่กำหนด"), so access ends at 00:00 of the next day in
// the Holding's timezone (authmodels.AccessEndsAt). Zero time = no expiry. Login and
// permission checks must all use this one rule: access is allowed only while now is
// before the returned instant (authmodels.AccessExpired).
func AccessExpiryInstant(expiryDate sql.NullTime, timezone string) time.Time {
	if !expiryDate.Valid {
		return time.Time{}
	}
	return authmodels.AccessEndsAt(expiryDate.Time, timezone)
}
