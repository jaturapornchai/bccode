package rolepermission

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/textguard"
)

// Permission-set rows are keyed by (holding_code, role_code), which is case-sensitive, and older
// data can hold one code in two spellings ("admin" and "ADMIN", sometimes with the same id). Every
// write therefore resolves the one stored row it means first and then touches only that row.
var (
	errPermissionSetNotFound  = errors.New("permission set not found")
	errPermissionSetAmbiguous = errors.New("permission set ambiguous")
	errPermissionSetCodeTaken = errors.New("permission set code exists")
	// errPermissionSetChanged: the update carried a __v older than the stored row (someone saved
	// it after this screen loaded it) — refused so the newer save is not overwritten.
	errPermissionSetChanged = errors.New("permission set changed meanwhile")
)

type permissionSetRow struct {
	id, roleCode string
	active       bool
}

// matchPermissionSets returns the rows whose id is ref or whose code equals ref ignoring case.
func matchPermissionSets(ctx context.Context, db *sql.DB, holding, ref string) ([]permissionSetRow, error) {
	rows, err := db.QueryContext(ctx, `SELECT COALESCE(id, ''), role_code, COALESCE(is_active, true)
		FROM role_permissions WHERE holding_code = $1 AND (id = $2 OR UPPER(role_code) = UPPER($2))
		ORDER BY role_code`, holding, ref)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := []permissionSetRow{}
	for rows.Next() {
		var row permissionSetRow
		if err := rows.Scan(&row.id, &row.roleCode, &row.active); err != nil {
			return nil, err
		}
		matches = append(matches, row)
	}
	return matches, rows.Err()
}

// resolvePermissionSet returns the stored role_code that ref (a row id or a code) names: the row
// with that id, else the row with exactly that code, else the only row whose code matches
// ignoring case. Two rows at the same step are ambiguous; nothing is guessed.
func resolvePermissionSet(ctx context.Context, db *sql.DB, holding, ref string) (string, error) {
	rows, err := matchPermissionSets(ctx, db, holding, ref)
	if err != nil {
		return "", err
	}
	for _, match := range []func(permissionSetRow) bool{
		func(row permissionSetRow) bool { return row.id != "" && row.id == ref },
		func(row permissionSetRow) bool { return row.roleCode == ref },
		func(row permissionSetRow) bool { return strings.EqualFold(row.roleCode, ref) },
	} {
		found := []string{}
		for _, row := range rows {
			if match(row) {
				found = append(found, row.roleCode)
			}
		}
		switch len(found) {
		case 0:
			continue
		case 1:
			return found[0], nil
		default:
			return "", errPermissionSetAmbiguous
		}
	}
	return "", errPermissionSetNotFound
}

// permissionSetValues are the columns a create or update writes.
type permissionSetValues struct {
	names, permissions []byte
	isActive           bool
}

// permissionSetSaved is what a create stored: the row id, stored code, real timestamps and version.
type permissionSetSaved struct {
	id, roleCode         string
	createdAt, updatedAt time.Time
	version              int64
}

// createPermissionSet adds roleCode (already upper-case) and returns what was stored (row id,
// stored code, timestamps, version). A code already used in any spelling is refused, except one
// deleted set with that code, which is brought back as the same row so members that still list
// it keep pointing at it.
func createPermissionSet(ctx context.Context, db *sql.DB, holding, roleCode string, values permissionSetValues) (permissionSetSaved, error) {
	rows, err := matchPermissionSets(ctx, db, holding, roleCode)
	if err != nil {
		return permissionSetSaved{}, err
	}
	same := []permissionSetRow{}
	for _, row := range rows {
		if !strings.EqualFold(row.roleCode, roleCode) {
			continue
		}
		if row.active {
			return permissionSetSaved{}, errPermissionSetCodeTaken
		}
		same = append(same, row)
	}
	newID := "rp-" + strings.ToLower(roleCode)
	saved := permissionSetSaved{roleCode: roleCode}
	switch len(same) {
	case 0:
		err = db.QueryRowContext(ctx, `INSERT INTO role_permissions (id, holding_code, role_code, names, permissions, is_active, created_at, updated_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, now(), now(), 0)
			RETURNING id, created_at, updated_at, version`, newID, holding, roleCode, values.names, values.permissions, values.isActive).
			Scan(&saved.id, &saved.createdAt, &saved.updatedAt, &saved.version)
		if centraldb.IsUniqueViolation(err) {
			return permissionSetSaved{}, errPermissionSetCodeTaken
		}
		return saved, err
	case 1:
		saved.roleCode = same[0].roleCode
		err = db.QueryRowContext(ctx, `UPDATE role_permissions SET id = COALESCE(id, $3), names = $4, permissions = $5,
				is_active = $6, updated_at = now(), version = version + 1
			WHERE holding_code = $1 AND role_code = $2 AND COALESCE(is_active, true) = false
			RETURNING id, COALESCE(created_at, updated_at), updated_at, version`, holding, same[0].roleCode, newID, values.names, values.permissions, values.isActive).
			Scan(&saved.id, &saved.createdAt, &saved.updatedAt, &saved.version)
		if errors.Is(err, sql.ErrNoRows) {
			return permissionSetSaved{}, errPermissionSetCodeTaken // brought back by someone else meanwhile
		}
		return saved, err
	default:
		return permissionSetSaved{}, errPermissionSetAmbiguous
	}
}

// updatePermissionSet rewrites the stored row when it still has version (the __v the screen
// loaded) and returns the new version; an older version is errPermissionSetChanged (409). A new
// code equal to the stored one ignoring case keeps the stored spelling; a real rename must not
// collide with another set in any spelling.
func updatePermissionSet(ctx context.Context, db *sql.DB, holding, stored, requested string, version int64, values permissionSetValues) (int64, error) {
	code := stored
	if requested != "" && !strings.EqualFold(requested, stored) {
		code = requested
		var taken bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM role_permissions
			WHERE holding_code = $1 AND UPPER(role_code) = UPPER($2) AND role_code <> $3)`, holding, code, stored).Scan(&taken); err != nil {
			return 0, err
		}
		if taken {
			return 0, errPermissionSetCodeTaken
		}
	}
	var next int64
	err := db.QueryRowContext(ctx, `UPDATE role_permissions SET role_code = $3, names = $4, permissions = $5,
			is_active = $6, updated_at = now(), version = version + 1
		WHERE holding_code = $1 AND role_code = $2 AND version = $7
		RETURNING version`, holding, stored, code, values.names, values.permissions, values.isActive, version).Scan(&next)
	if centraldb.IsUniqueViolation(err) {
		return 0, errPermissionSetCodeTaken
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, versionMissed(ctx, db, holding, stored)
	}
	return next, err
}

// deletePermissionSet switches the stored row off (members keep their list; the set grants nothing)
// when it still has version (the __v the screen loaded). An older version is errPermissionSetChanged:
// a delete confirmed on a stale list must not switch off a set someone re-enabled or re-granted
// meanwhile (review 2026-09-24 — the delete ignored the __v the screen sends).
func deletePermissionSet(ctx context.Context, db *sql.DB, holding, stored string, version int64) error {
	var next int64
	err := db.QueryRowContext(ctx, `UPDATE role_permissions SET is_active = false, updated_at = now(), version = version + 1
		WHERE holding_code = $1 AND role_code = $2 AND version = $3
		RETURNING version`, holding, stored, version).Scan(&next)
	if errors.Is(err, sql.ErrNoRows) {
		return versionMissed(ctx, db, holding, stored)
	}
	return err
}

// versionMissed tells why a version-locked write touched no row: the row is still there with a
// newer version (errPermissionSetChanged, 409) or it was removed meanwhile (errPermissionSetNotFound).
func versionMissed(ctx context.Context, db *sql.DB, holding, stored string) error {
	var exists bool
	if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM role_permissions WHERE holding_code = $1 AND role_code = $2)`,
		holding, stored).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return errPermissionSetChanged
	}
	return errPermissionSetNotFound
}

// respondPermissionSetError answers a failed permission-set write, naming the field to fix.
func respondPermissionSetError(ctx microservice.IContext, err error) error {
	var base *apperr.AppError
	key, field := "", ""
	switch {
	case errors.Is(err, errPermissionSetCodeTaken):
		base, key, field = apperr.ErrDuplicate, "ss_err_permission_set_code_exists", "rolecode"
	case errors.Is(err, errPermissionSetAmbiguous):
		base, key, field = apperr.ErrConflict, "ss_err_permission_set_ambiguous", "rolecode"
	case errors.Is(err, errPermissionSetChanged):
		base, key, field = apperr.ErrConflict, "ss_err_permission_set_changed", "__v"
	case errors.Is(err, errPermissionSetNotFound):
		base, key = apperr.ErrNotFound, "ss_err_not_found"
	default:
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	lang := requestLanguage(ctx)
	appErr := base.WithMessage(language.Text(key, lang)).WithThaiMessage(language.Text(key, "th")).WithWrap(err)
	if field != "" {
		appErr = appErr.WithField(field)
	}
	return apperr.Respond(ctx, appErr)
}

// permissionSetNULLabels names the permission set fields (languages.tsv keys) in the NUL message.
var permissionSetNULLabels = map[string]string{"rolecode": "ss_f_permission_set_code", "name": "ss_f_permission_set_name"}

// respondPermissionSetNUL answers a write whose text carries U+0000 (usually pasted from a PDF):
// PostgreSQL JSONB cannot store it, so the save failed as a generic 500 that named no field.
func respondPermissionSetNUL(ctx microservice.IContext, path string) error {
	message := func(lang string) string {
		return textguard.Message(path, permissionSetNULLabels, func(key string) string { return language.Text(key, lang) })
	}
	return apperr.Respond(ctx, apperr.ErrValidation.WithMessage(message(requestLanguage(ctx))).WithThaiMessage(message("th")).WithField(path))
}

// requestLanguage is the caller's language: ?lang= first, else Accept-Language.
func requestLanguage(ctx microservice.IContext) string {
	if lang := strings.TrimSpace(ctx.QueryParam("lang")); lang != "" {
		return lang
	}
	return ctx.Header("Accept-Language")
}
