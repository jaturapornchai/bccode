package businesstype

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"

	"smlcloudplatform/internal/centraldb"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/textguard"
)

// business_types is unique on the exact (holding_code, code), so "retail" and "RETAIL" could both
// be stored. Creates refuse a code already used in any spelling, and updates/deletes resolve the
// one row they mean before touching it.
var (
	errBusinessTypeNotFound  = errors.New("business type not found")
	errBusinessTypeAmbiguous = errors.New("business type ambiguous")
	errBusinessTypeCodeTaken = errors.New("business type code exists")
)

type businessTypeRow struct {
	id, code string
	active   bool
}

// matchBusinessTypes returns the rows whose id is ref or whose code equals ref ignoring case.
func matchBusinessTypes(ctx context.Context, db *sql.DB, holding, ref string) ([]businessTypeRow, error) {
	rows, err := db.QueryContext(ctx, `SELECT id, code, is_active FROM business_types
		WHERE holding_code = $1 AND (id = $2 OR UPPER(code) = UPPER($2)) ORDER BY code`, holding, ref)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	matches := []businessTypeRow{}
	for rows.Next() {
		var row businessTypeRow
		if err := rows.Scan(&row.id, &row.code, &row.active); err != nil {
			return nil, err
		}
		matches = append(matches, row)
	}
	return matches, rows.Err()
}

// resolveBusinessType returns the id of the row ref (an id or a code) names: the row with that id,
// else the row with exactly that code, else the only row whose code matches ignoring case.
func resolveBusinessType(ctx context.Context, db *sql.DB, holding, ref string) (string, error) {
	rows, err := matchBusinessTypes(ctx, db, holding, ref)
	if err != nil {
		return "", err
	}
	for _, match := range []func(businessTypeRow) bool{
		func(row businessTypeRow) bool { return row.id == ref },
		func(row businessTypeRow) bool { return row.code == ref },
		func(row businessTypeRow) bool { return strings.EqualFold(row.code, ref) },
	} {
		found := []string{}
		for _, row := range rows {
			if match(row) {
				found = append(found, row.id)
			}
		}
		switch len(found) {
		case 0:
			continue
		case 1:
			return found[0], nil
		default:
			return "", errBusinessTypeAmbiguous
		}
	}
	return "", errBusinessTypeNotFound
}

// createBusinessType adds code and returns the row id. A code already used in
// any spelling is refused, except one deleted type with that code, which is brought back as the
// same row (same id, so branches that still point at it stay valid).
func createBusinessType(ctx context.Context, db *sql.DB, holding, code string, names []byte, isDefault bool) (string, error) {
	rows, err := matchBusinessTypes(ctx, db, holding, code)
	if err != nil {
		return "", err
	}
	deleted := []businessTypeRow{}
	for _, row := range rows {
		if !strings.EqualFold(row.code, code) {
			continue
		}
		if row.active {
			return "", errBusinessTypeCodeTaken
		}
		deleted = append(deleted, row)
	}
	switch len(deleted) {
	case 0:
		id := uuid.NewString()
		_, err = db.ExecContext(ctx, `INSERT INTO business_types (id, holding_code, code, names, is_default, is_active)
			VALUES ($1, $2, $3, $4, $5, true)`, id, holding, code, names, isDefault)
		if centraldb.IsUniqueViolation(err) {
			return "", errBusinessTypeCodeTaken
		}
		return id, err
	case 1:
		result, err := db.ExecContext(ctx, `UPDATE business_types SET names = $3, is_default = $4, is_active = true, updated_at = now()
			WHERE holding_code = $1 AND id = $2 AND is_active = false`, holding, deleted[0].id, names, isDefault)
		if err = oneBusinessTypeChanged(result, err); errors.Is(err, errBusinessTypeNotFound) {
			return "", errBusinessTypeCodeTaken // brought back by someone else meanwhile
		}
		return deleted[0].id, err
	default:
		return "", errBusinessTypeAmbiguous
	}
}

func oneBusinessTypeChanged(result sql.Result, err error) error {
	if err != nil {
		return err
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if changed != 1 {
		return errBusinessTypeNotFound
	}
	return nil
}

// businessTypeNULLabels names the business type fields (languages.tsv keys) in the NUL message.
var businessTypeNULLabels = map[string]string{"code": "business_type_code", "name": "business_type_name"}

// respondBusinessTypeNUL answers a write whose text carries U+0000 (usually pasted from a PDF):
// PostgreSQL cannot store it, so the INSERT failed as a generic 500 that named no field.
func respondBusinessTypeNUL(ctx microservice.IContext, path string) error {
	lang := strings.TrimSpace(ctx.QueryParam("lang"))
	if lang == "" {
		lang = ctx.Header("Accept-Language")
	}
	message := func(lang string) string {
		return textguard.Message(path, businessTypeNULLabels, func(key string) string { return language.Text(key, lang) })
	}
	return apperr.Respond(ctx, apperr.ErrValidation.WithMessage(message(lang)).WithThaiMessage(message("th")).WithField(path))
}

// respondBusinessTypeError answers a failed business-type write, naming the field to fix.
func respondBusinessTypeError(ctx microservice.IContext, err error) error {
	var base *apperr.AppError
	key, field := "", ""
	switch {
	case errors.Is(err, errBusinessTypeCodeTaken):
		base, key, field = apperr.ErrDuplicate, "ss_err_business_type_code_exists", "code"
	case errors.Is(err, errBusinessTypeAmbiguous):
		base, key, field = apperr.ErrConflict, "ss_err_business_type_ambiguous", "code"
	case errors.Is(err, errBusinessTypeNotFound):
		base, key = apperr.ErrNotFound, "ss_err_not_found"
	default:
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	lang := strings.TrimSpace(ctx.QueryParam("lang"))
	if lang == "" {
		lang = ctx.Header("Accept-Language")
	}
	appErr := base.WithMessage(language.Text(key, lang)).WithThaiMessage(language.Text(key, "th")).WithWrap(err)
	if field != "" {
		appErr = appErr.WithField(field)
	}
	return apperr.Respond(ctx, appErr)
}
